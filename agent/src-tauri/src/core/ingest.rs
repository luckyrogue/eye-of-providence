// Ingest pump: периодически берёт события из LocalStore и шлёт в backend.
//
// Retry-policy:
//   - 5xx / network error / 429: exponential backoff (1s → 2s → 4s → … cap 5min)
//     с release lease (события остаются в очереди).
//   - 401: помечаем токен невалидным, эмитим tauri event `auth-required`,
//     release lease, останавливаем активность до wake (например, после
//     успешного pair_poll).
//   - 4xx (≠401/429): drop batch (commit) — данные битые, retry бесполезен.
//   - 2xx: commit.
//
// Креды (URL + token) читаются провайдерами на каждом tick, поэтому смена
// токена через pairing подхватывается без рестарта pump'а.

use anyhow::Result;
use serde::Serialize;
use std::sync::Arc;
use std::time::Duration;
use tokio::sync::Notify;

use super::event::Event;
use super::store::LocalStore;

#[derive(Serialize)]
struct IngestRequest<'a> {
    events: &'a [Event],
}

/// Сигнал из UI: токен потенциально обновился — разбудить loop.
#[derive(Clone, Default)]
pub struct WakeUp(Arc<Notify>);

impl WakeUp {
    pub fn notify(&self) {
        self.0.notify_waiters();
    }
    pub async fn notified(&self) {
        self.0.notified().await;
    }
}

/// Провайдер кредов: возвращает (backend_url, token). Любое поле может быть
/// None — тогда tick пропускает iteration.
pub type CredsProvider = Arc<dyn Fn() -> (Option<String>, Option<String>) + Send + Sync>;

pub struct AuthEmitter(Box<dyn Fn() + Send + Sync + 'static>);

impl AuthEmitter {
    pub fn new<F: Fn() + Send + Sync + 'static>(f: F) -> Self {
        Self(Box::new(f))
    }
    fn emit(&self) {
        (self.0)();
    }
}

pub struct Ingest {
    store: Arc<LocalStore>,
    creds: CredsProvider,
    batch_size: usize,
    flush_interval: Duration,
    http: reqwest::Client,
    emitter: AuthEmitter,
    wake: WakeUp,
}

impl Ingest {
    pub fn new(
        store: Arc<LocalStore>,
        creds: CredsProvider,
        batch_size: usize,
        flush_interval: Duration,
        emitter: AuthEmitter,
        wake: WakeUp,
    ) -> Self {
        Self {
            store,
            creds,
            batch_size,
            flush_interval,
            http: reqwest::Client::builder()
                .timeout(Duration::from_secs(10))
                .build()
                .expect("reqwest client"),
            emitter,
            wake,
        }
    }

    pub fn spawn(self) -> tauri::async_runtime::JoinHandle<()> {
        tauri::async_runtime::spawn(async move {
            let mut backoff = Duration::from_secs(1);
            let max_backoff = Duration::from_secs(5 * 60);
            loop {
                match self.tick().await {
                    Ok(TickOutcome::Ok) | Ok(TickOutcome::Empty) => {
                        backoff = Duration::from_secs(1);
                        tokio::select! {
                            _ = tokio::time::sleep(self.flush_interval) => {}
                            _ = self.wake.notified() => {}
                        }
                    }
                    Ok(TickOutcome::Retryable) => {
                        tracing::warn!(secs = backoff.as_secs(), "ingest retry backoff");
                        tokio::select! {
                            _ = tokio::time::sleep(backoff) => {}
                            _ = self.wake.notified() => { backoff = Duration::from_secs(1); }
                        }
                        backoff = (backoff * 2).min(max_backoff);
                    }
                    Ok(TickOutcome::NoCreds) => {
                        tracing::debug!("no creds, waiting for pairing");
                        self.wake.notified().await;
                    }
                    Ok(TickOutcome::AuthRequired) => {
                        tracing::warn!("ingest paused: auth required");
                        self.wake.notified().await;
                        backoff = Duration::from_secs(1);
                    }
                    Err(err) => {
                        tracing::warn!(error = %err, "ingest tick error");
                        tokio::time::sleep(backoff).await;
                        backoff = (backoff * 2).min(max_backoff);
                    }
                }
            }
        })
    }

    async fn tick(&self) -> Result<TickOutcome> {
        let (backend, token) = (self.creds)();
        let (backend, token) = match (backend, token) {
            (Some(b), Some(t)) if !b.is_empty() && !t.is_empty() => (b, t),
            _ => return Ok(TickOutcome::NoCreds),
        };

        let leased = self.store.lease_batch(self.batch_size, 60)?;
        if leased.is_empty() {
            return Ok(TickOutcome::Empty);
        }
        let ids: Vec<i64> = leased.iter().map(|(id, _)| *id).collect();
        let events: Vec<Event> = leased.into_iter().map(|(_, e)| e).collect();

        let url = format!("{}/v1/ingest", backend.trim_end_matches('/'));
        let resp = self
            .http
            .post(url)
            .bearer_auth(&token)
            .json(&IngestRequest { events: &events })
            .send()
            .await;

        let result = match resp {
            Ok(r) if r.status().is_success() => {
                self.store.commit(&ids)?;
                tracing::debug!(count = ids.len(), "ingest commit");
                TickOutcome::Ok
            }
            Ok(r) if r.status() == reqwest::StatusCode::UNAUTHORIZED => {
                tracing::warn!("ingest 401 — токен невалиден, требуется re-pair");
                self.store.release(&ids)?;
                self.emitter.emit();
                TickOutcome::AuthRequired
            }
            Ok(r) if r.status() == reqwest::StatusCode::TOO_MANY_REQUESTS => {
                self.store.release(&ids)?;
                TickOutcome::Retryable
            }
            Ok(r) if r.status().is_server_error() => {
                tracing::warn!(status = ?r.status(), "ingest 5xx, retryable");
                self.store.release(&ids)?;
                TickOutcome::Retryable
            }
            Ok(r) => {
                tracing::warn!(
                    status = ?r.status(),
                    count = ids.len(),
                    "ingest 4xx — drop batch"
                );
                self.store.commit(&ids)?;
                TickOutcome::Ok
            }
            Err(err) => {
                tracing::warn!(error = %err, "ingest network error, retryable");
                self.store.release(&ids)?;
                TickOutcome::Retryable
            }
        };
        Ok(result)
    }
}

#[derive(Debug, Clone, Copy)]
enum TickOutcome {
    Ok,
    Empty,
    NoCreds,
    Retryable,
    AuthRequired,
}
