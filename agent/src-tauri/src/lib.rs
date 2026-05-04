mod core;
mod platform;

use tracing_subscriber::EnvFilter;

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")))
        .init();

    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![status])
        .setup(|_app| {
            tracing::info!("eop agent started");
            // TODO Phase 1: spawn platform watcher, init SQLite buffer, start ingest pump.
            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[tauri::command]
fn status() -> serde_json::Value {
    serde_json::json!({
        "status": "ok",
        "phase": "0-skeleton"
    })
}
