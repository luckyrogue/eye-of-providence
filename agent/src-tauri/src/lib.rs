mod core;
mod platform;

use std::sync::Arc;
use std::time::Duration;

use tauri::menu::{Menu, MenuItem};
use tauri::tray::TrayIconBuilder;
use tauri::Manager;
use tracing_subscriber::EnvFilter;

use crate::core::ingest::{Ingest, IngestConfig};
use crate::core::local_api;
use crate::core::preflight::{self, CheckResult, PreflightInput};
use crate::core::store::LocalStore;
use crate::core::watcher;

struct AgentState {
    store: Arc<LocalStore>,
    data_dir: std::path::PathBuf,
    local_api_port: u16,
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    tracing_subscriber::fmt()
        .with_env_filter(EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info")))
        .init();

    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![
            status,
            pending_count,
            check_accessibility,
            preflight_run,
        ])
        .setup(|app| {
            // Tray
            let show = MenuItem::with_id(app, "show", "Open dashboard", true, None::<&str>)?;
            let pause = MenuItem::with_id(app, "pause", "Pause tracking", true, None::<&str>)?;
            let quit = MenuItem::with_id(app, "quit", "Quit", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show, &pause, &quit])?;
            let _tray = TrayIconBuilder::new()
                .menu(&menu)
                .tooltip("Eye of Providence")
                .on_menu_event(|app, event| match event.id.as_ref() {
                    "show" => {
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                    "pause" => {
                        tracing::info!("pause tracking (todo: actual toggle)");
                    }
                    "quit" => app.exit(0),
                    _ => {}
                })
                .build(app)?;

            let data_dir = app.path().app_data_dir()?;
            std::fs::create_dir_all(&data_dir)?;
            let db_path = data_dir.join("eop.sqlite");
            tracing::info!(path = %db_path.display(), "opening local store");

            let store = Arc::new(LocalStore::open(&db_path)?);
            let local_api_port: u16 = std::env::var("EOP_LOCAL_API_PORT")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(7373);
            app.manage(AgentState {
                store: store.clone(),
                data_dir: data_dir.clone(),
                local_api_port,
            });

            // Preflight на старте — пишем в лог, дальше UI запросит сам через команду.
            let pf_dir = data_dir.clone();
            tauri::async_runtime::spawn(async move {
                let backend = std::env::var("EOP_BACKEND_URL").ok();
                let token = std::env::var("EOP_BEARER_TOKEN").ok();
                let results = preflight::run(PreflightInput {
                    data_dir: &pf_dir,
                    local_api_port,
                    check_local_api_port: false,
                    backend_url: backend.as_deref(),
                    bearer_token: token.as_deref(),
                }).await;
                for r in &results {
                    match r.status {
                        preflight::CheckStatus::Ok => tracing::info!(check = %r.id, "{}", r.message),
                        preflight::CheckStatus::Warn => tracing::warn!(check = %r.id, "{}", r.message),
                        preflight::CheckStatus::Error => tracing::error!(check = %r.id, "{}", r.message),
                    }
                }
            });

            // Local API для browser extension и IDE plugin.
            // Token живёт в data_dir, browser-extension читает его через user-paste.
            let token_path = data_dir.join("eop.local-token");
            let token = if token_path.exists() {
                std::fs::read_to_string(&token_path)?.trim().to_string()
            } else {
                let new_tok: String = (0..32)
                    .map(|_| {
                        let n: u8 = rand_byte() % 36;
                        if n < 10 {
                            (b'0' + n) as char
                        } else {
                            (b'a' + (n - 10)) as char
                        }
                    })
                    .collect();
                std::fs::write(&token_path, &new_tok)?;
                new_tok
            };
            local_api::spawn(store.clone(), token, local_api_port);

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
        "phase": "8-final"
    })
}

#[tauri::command]
fn pending_count(state: tauri::State<'_, AgentState>) -> Result<i64, String> {
    state.store.pending_count().map_err(|e| e.to_string())
}

/// Возвращает true, если у приложения есть Accessibility permission (macOS),
/// или true для остальных платформ (Windows/Linux не требуют этого permission).
#[tauri::command]
fn check_accessibility() -> bool {
    #[cfg(target_os = "macos")]
    {
        return platform::macos::has_accessibility();
    }
    #[cfg(not(target_os = "macos"))]
    {
        true
    }
}

/// Запустить preflight по требованию UI (Onboarding screen).
#[tauri::command]
async fn preflight_run(state: tauri::State<'_, AgentState>) -> Result<Vec<CheckResult>, String> {
    let backend = std::env::var("EOP_BACKEND_URL").ok();
    let token = std::env::var("EOP_BEARER_TOKEN").ok();
    Ok(preflight::run(PreflightInput {
        data_dir: &state.data_dir,
        local_api_port: state.local_api_port,
        check_local_api_port: true,
        backend_url: backend.as_deref(),
        bearer_token: token.as_deref(),
    }).await)
}

// Простейший источник «псевдослучайных» байт для local-token (не cryptographic;
// токен живёт только локально на машине пользователя). Используем системные nanoseconds.
fn rand_byte() -> u8 {
    use std::time::{SystemTime, UNIX_EPOCH};
    let n = SystemTime::now().duration_since(UNIX_EPOCH).map(|d| d.subsec_nanos()).unwrap_or(0);
    (n & 0xff) as u8
}
