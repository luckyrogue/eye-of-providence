// Ingest pump: периодически берёт события из LocalStore и шлёт в backend.

use anyhow::Result;
use serde::Serialize;
use std::sync::Arc;
use std::time::Duration;

use super::event::Event;
use super::store::LocalStore;

#[derive(Clone)]
pub struct IngestConfig {
    pub base_url: String,
    pub bearer_token: String,
    pub batch_size: usize,
    pub flush_interval: Duration,
}

#[derive(Serialize)]
struct IngestRequest<'a> {
    events: &'a [Event],
}

pub struct Ingest {
    store: Arc<LocalStore>,
    cfg: IngestConfig,
    http: reqwest::Client,
}

impl Ingest {
    pub fn new(store: Arc<LocalStore>, cfg: IngestConfig) -> Self {
        Self {
            store,
            cfg,
            http: reqwest::Client::builder()
                .timeout(Duration::from_secs(10))
                .build()
                .expect("reqwest client"),
        }
    }

    /// Запускает фоновый цикл: каждые flush_interval вытаскивает batch и шлёт.
    pub fn spawn(self) -> tauri::async_runtime::JoinHandle<()> {
        tauri::async_runtime::spawn(async move {
            loop {
                if let Err(err) = self.tick().await {
                    tracing::warn!(error = %err, "ingest tick failed");
                }
                tokio::time::sleep(self.cfg.flush_interval).await;
            }
        })
    }

    async fn tick(&self) -> Result<()> {
        let leased = self.store.lease_batch(self.cfg.batch_size, 60)?;
        if leased.is_empty() {
            return Ok(());
        }
        let ids: Vec<i64> = leased.iter().map(|(id, _)| *id).collect();
        let events: Vec<Event> = leased.into_iter().map(|(_, e)| e).collect();

        let url = format!("{}/v1/ingest", self.cfg.base_url.trim_end_matches('/'));
        let resp = self
            .http
            .post(url)
            .bearer_auth(&self.cfg.bearer_token)
            .json(&IngestRequest { events: &events })
            .send()
            .await;

        match resp {
            Ok(r) if r.status().is_success() => {
                self.store.commit(&ids)?;
                tracing::debug!(count = ids.len(), "ingest commit");
            }
            Ok(r) => {
                tracing::warn!(status = ?r.status(), "ingest non-2xx, releasing lease");
                self.store.release(&ids)?;
            }
            Err(err) => {
                tracing::warn!(error = %err, "ingest network error, releasing lease");
                self.store.release(&ids)?;
            }
        }
        Ok(())
    }
}
