//! Auth для desktop-агента.
//!
//! Хранилище: `keyring` крейт даёт нативный backend на каждой ОС
//!   - macOS  → Keychain
//!   - Windows → Credential Manager
//!   - Linux  → secret-service (gnome-keyring / kwallet)
//!
//! Что храним:
//!   - token       — opaque eop_*** API token (scope=write:ingest, kind=agent)
//!   - backend_url — URL без хвостового `/api/` (нормализуем при чтении)
//!   - user_id     — uuid для отображения в Settings
//!
//! Pairing flow реализован в tauri-командах `pair_begin` / `pair_poll`.

use anyhow::{anyhow, Context, Result};
use keyring::Entry;
use serde::{Deserialize, Serialize};

const SERVICE: &str = "eop-agent";
const ACCOUNT_TOKEN: &str = "ingest-token";
const ACCOUNT_BACKEND: &str = "backend-url";
const ACCOUNT_USER: &str = "user-id";

pub const DEFAULT_BACKEND: &str = "https://eop.rysdavletov.org/api";

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AccountInfo {
    pub user_id: Option<String>,
    pub backend_url: String,
    pub has_token: bool,
}

fn entry(account: &str) -> Result<Entry> {
    Entry::new(SERVICE, account).map_err(|e| anyhow!("keyring entry: {e}"))
}

fn read_opt(account: &str) -> Result<Option<String>> {
    match entry(account)?.get_password() {
        Ok(v) => Ok(Some(v)),
        Err(keyring::Error::NoEntry) => Ok(None),
        Err(e) => Err(anyhow!("keyring read {account}: {e}")),
    }
}

fn write(account: &str, value: &str) -> Result<()> {
    entry(account)?
        .set_password(value)
        .with_context(|| format!("keyring write {account}"))
}

fn delete(account: &str) -> Result<()> {
    match entry(account)?.delete_credential() {
        Ok(_) | Err(keyring::Error::NoEntry) => Ok(()),
        Err(e) => Err(anyhow!("keyring delete {account}: {e}")),
    }
}

/// Прочесть токен (если связан). Возвращает `Ok(None)` если не привязан.
pub fn get_token() -> Result<Option<String>> {
    read_opt(ACCOUNT_TOKEN)
}

/// URL бэкенда. Fallback на env (`EOP_BACKEND_URL`) для dev, дальше — DEFAULT.
pub fn get_backend_url() -> Result<String> {
    if let Some(v) = read_opt(ACCOUNT_BACKEND)? {
        return Ok(v);
    }
    if let Ok(v) = std::env::var("EOP_BACKEND_URL") {
        if !v.trim().is_empty() {
            return Ok(v);
        }
    }
    Ok(DEFAULT_BACKEND.to_string())
}

pub fn set_backend_url(url: &str) -> Result<()> {
    write(ACCOUNT_BACKEND, url.trim_end_matches('/'))
}

pub fn account_info() -> Result<AccountInfo> {
    Ok(AccountInfo {
        user_id: read_opt(ACCOUNT_USER)?,
        backend_url: get_backend_url()?,
        has_token: get_token()?.is_some(),
    })
}

pub fn logout() -> Result<()> {
    delete(ACCOUNT_TOKEN)?;
    delete(ACCOUNT_USER)?;
    Ok(())
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PairBeginResponse {
    pub pair_id: String,
    pub secret: String,
    pub code: String,
    pub expires_in: i32,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct PollResponse {
    pub status: String,
    pub token: Option<String>,
    pub user_id: Option<String>,
}

pub async fn pair_begin(http: &reqwest::Client, backend: &str) -> Result<PairBeginResponse> {
    let url = format!("{}/v1/devices/pair", backend.trim_end_matches('/'));
    let body = serde_json::json!({ "kind": "agent", "name": format!("{} agent", hostname()) });
    let resp = http
        .post(&url)
        .json(&body)
        .send()
        .await
        .with_context(|| format!("pair_begin POST {url}"))?;
    if !resp.status().is_success() {
        return Err(anyhow!("pair_begin status {}", resp.status()));
    }
    Ok(resp.json::<PairBeginResponse>().await?)
}

pub async fn pair_poll(
    http: &reqwest::Client,
    backend: &str,
    pair_id: &str,
    secret: &str,
) -> Result<PollResponse> {
    let url = format!("{}/v1/devices/poll", backend.trim_end_matches('/'));
    let body = serde_json::json!({ "pair_id": pair_id, "secret": secret });
    let resp = http
        .post(&url)
        .json(&body)
        .send()
        .await
        .with_context(|| format!("pair_poll POST {url}"))?;
    if resp.status() == reqwest::StatusCode::NOT_FOUND {
        return Ok(PollResponse {
            status: "expired".to_string(),
            token: None,
            user_id: None,
        });
    }
    if !resp.status().is_success() {
        return Err(anyhow!("pair_poll status {}", resp.status()));
    }
    Ok(resp.json::<PollResponse>().await?)
}

/// Сохранить полученный после claim token + user_id в keyring.
pub fn persist_pairing(token: &str, user_id: &str) -> Result<()> {
    write(ACCOUNT_TOKEN, token)?;
    write(ACCOUNT_USER, user_id)?;
    Ok(())
}

fn hostname() -> String {
    std::env::var("HOST").unwrap_or_else(|_| {
        std::env::var("COMPUTERNAME").unwrap_or_else(|_| "desktop".to_string())
    })
}
