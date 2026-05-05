// Локальный HTTP API на 127.0.0.1, через который browser extension и
// VS Code plugin могут поставлять события в общий буфер агента,
// не зная JWT для backend (агент сам добавит при отправке).
//
// Pairing: каждый клиент должен указать `Authorization: Bearer <token>`,
// где <token> — содержимое файла ~/<data>/eop.local-token (создаётся при
// первом запуске). UI показывает токен в окне Settings.

use anyhow::Result;
use serde::Deserialize;
use std::sync::Arc;
use tokio::io::{AsyncBufReadExt, AsyncReadExt, AsyncWriteExt, BufReader};
use tokio::net::{TcpListener, TcpStream};

use super::event::Event;
use super::store::LocalStore;

#[derive(Deserialize)]
struct IngestBody {
    events: Vec<Event>,
}

pub fn spawn(store: Arc<LocalStore>, token: String, port: u16) -> tokio::task::JoinHandle<()> {
    tokio::spawn(async move {
        let addr = format!("127.0.0.1:{}", port);
        let listener = match TcpListener::bind(&addr).await {
            Ok(l) => l,
            Err(err) => {
                tracing::warn!(error = %err, addr = %addr, "local API: bind failed");
                return;
            }
        };
        tracing::info!(%addr, "local API listening");

        loop {
            let (stream, _) = match listener.accept().await {
                Ok(v) => v,
                Err(err) => {
                    tracing::warn!(error = %err, "accept failed");
                    continue;
                }
            };
            let store = store.clone();
            let token = token.clone();
            tokio::spawn(async move {
                if let Err(err) = handle(stream, store, token).await {
                    tracing::debug!(error = %err, "local conn ended");
                }
            });
        }
    })
}

async fn handle(mut stream: TcpStream, store: Arc<LocalStore>, expected_token: String) -> Result<()> {
    let (read, mut write) = stream.split();
    let mut reader = BufReader::new(read);

    // Читаем request line
    let mut line = String::new();
    reader.read_line(&mut line).await?;
    let (method, path) = {
        let (m, p) = parse_request_line(&line);
        (m.to_string(), p.to_string())
    };

    // Читаем headers
    let mut content_length = 0usize;
    let mut auth_header = String::new();
    loop {
        line.clear();
        reader.read_line(&mut line).await?;
        if line == "\r\n" || line.is_empty() {
            break;
        }
        let lower = line.to_ascii_lowercase();
        if let Some(v) = lower.strip_prefix("content-length:") {
            content_length = v.trim().parse().unwrap_or(0);
        } else if let Some(v) = lower.strip_prefix("authorization:") {
            auth_header = v.trim().to_string();
        }
    }

    if !auth_header.eq_ignore_ascii_case(&format!("bearer {}", expected_token)) {
        return reply(&mut write, 401, b"{\"error\":\"unauthorized\"}").await;
    }

    if method != "POST" || path != "/v1/local/ingest" {
        return reply(&mut write, 404, b"{\"error\":\"not found\"}").await;
    }

    let mut body = vec![0u8; content_length];
    reader.read_exact(&mut body).await?;
    let parsed: IngestBody = match serde_json::from_slice(&body) {
        Ok(v) => v,
        Err(_) => return reply(&mut write, 400, b"{\"error\":\"bad json\"}").await,
    };

    for ev in &parsed.events {
        if let Err(err) = store.push(ev) {
            tracing::warn!(error = %err, "push failed");
        }
    }
    reply(&mut write, 200, format!("{{\"accepted\":{}}}", parsed.events.len()).as_bytes()).await
}

fn parse_request_line(line: &str) -> (&str, &str) {
    let mut parts = line.split_whitespace();
    let method = parts.next().unwrap_or("");
    let path = parts.next().unwrap_or("");
    (method, path)
}

async fn reply<W: AsyncWriteExt + Unpin>(w: &mut W, status: u16, body: &[u8]) -> Result<()> {
    let head = format!(
        "HTTP/1.1 {} OK\r\nContent-Type: application/json\r\nContent-Length: {}\r\nConnection: close\r\n\r\n",
        status,
        body.len()
    );
    w.write_all(head.as_bytes()).await?;
    w.write_all(body).await?;
    w.flush().await?;
    Ok(())
}
