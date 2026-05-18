// Автоматическая проверка локального API (порт слушает) и доступности backend.

use serde::Serialize;
use std::time::Duration;

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum BackendState {
    Online,
    Offline,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum LocalApiState {
    Online,
    Offline,
}

#[derive(Debug, Clone, Serialize)]
pub struct ConnectionStatus {
    pub backend: BackendState,
    pub local_api: LocalApiState,
    pub local_api_port: u16,
    pub paired: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub backend_url: Option<String>,
}

pub struct ConnectionInput<'a> {
    pub local_api_port: u16,
    pub http: &'a reqwest::Client,
}

pub async fn status(input: ConnectionInput<'_>) -> ConnectionStatus {
    let backend_url = crate::core::auth::get_backend_url().ok();
    let paired = crate::core::auth::get_token().ok().flatten().is_some();

    let (backend, backend_url) = match backend_url {
        Some(url) if !url.trim().is_empty() => {
            let state = probe_backend(input.http, &url).await;
            (state, Some(url))
        }
        _ => (BackendState::Offline, None),
    };

    let local_api = probe_local_api(input.local_api_port).await;

    ConnectionStatus {
        backend,
        local_api,
        local_api_port: input.local_api_port,
        paired,
        backend_url,
    }
}

async fn probe_backend(client: &reqwest::Client, url: &str) -> BackendState {
    let healthz = format!("{}/healthz", url.trim_end_matches('/'));
    match client
        .get(&healthz)
        .timeout(Duration::from_secs(3))
        .send()
        .await
    {
        Ok(r) if r.status().is_success() => BackendState::Online,
        _ => BackendState::Offline,
    }
}

async fn probe_local_api(port: u16) -> LocalApiState {
    let addr = format!("127.0.0.1:{}", port);
    match tokio::time::timeout(
        Duration::from_millis(500),
        tokio::net::TcpStream::connect(&addr),
    )
    .await
    {
        Ok(Ok(_)) => LocalApiState::Online,
        _ => LocalApiState::Offline,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[tokio::test]
    async fn local_api_offline_on_unused_port() {
        let port = {
            let listener = std::net::TcpListener::bind("127.0.0.1:0").unwrap();
            listener.local_addr().unwrap().port()
        };
        assert_eq!(probe_local_api(port).await, LocalApiState::Offline);
    }

    #[tokio::test]
    async fn local_api_online_when_listening() {
        let listener = tokio::net::TcpListener::bind("127.0.0.1:0").await.unwrap();
        let port = listener.local_addr().unwrap().port();
        assert_eq!(probe_local_api(port).await, LocalApiState::Online);
    }
}
