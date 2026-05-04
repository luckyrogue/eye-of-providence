mod core;
mod platform;

use std::sync::Arc;
use std::time::Duration;

use tauri::Manager;
use tracing_subscriber::EnvFilter;

use crate::core::ingest::{Ingest, IngestConfig};
use crate::core::store::LocalStore;
use crate::core::watcher;

struct AgentState {
    store: Arc<LocalStore>,
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")))
        .init();

    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![status, pending_count])
        .setup(|app| {
            let data_dir = app.path().app_data_dir()?;
            std::fs::create_dir_all(&data_dir)?;
            let db_path = data_dir.join("eop.sqlite");
            tracing::info!(path = %db_path.display(), "opening local store");

            let store = Arc::new(LocalStore::open(&db_path)?);
            app.manage(AgentState { store: store.clone() });

            // Watcher loop — каждые 5с опрашивает PlatformWatcher.
            let platform = platform::build();
            watcher::spawn(store.clone(), platform, Duration::from_secs(5), 90);

            // Ingest pump (только если есть конфиг).
            if let Ok(base_url) = std::env::var("EOP_BACKEND_URL") {
                if let Ok(token) = std::env::var("EOP_BEARER_TOKEN") {
                    let cfg = IngestConfig {
                        base_url,
                        bearer_token: token,
                        batch_size: 100,
                        flush_interval: Duration::from_secs(15),
                    };
                    Ingest::new(store, cfg).spawn();
                    tracing::info!("ingest pump started");
                } else {
                    tracing::warn!("EOP_BEARER_TOKEN not set — events будут копиться локально");
                }
            } else {
                tracing::info!("EOP_BACKEND_URL не задан — local-only mode");
            }

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[tauri::command]
fn status() -> serde_json::Value {
    serde_json::json!({
        "status": "ok",
        "phase": "1-skeleton-e2e"
    })
}

#[tauri::command]
fn pending_count(state: tauri::State<'_, AgentState>) -> Result<i64, String> {
    state.store.pending_count().map_err(|e| e.to_string())
}
