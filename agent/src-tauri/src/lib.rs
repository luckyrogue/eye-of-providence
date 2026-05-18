mod core;
mod platform;

use std::sync::Arc;
use std::time::Duration;

use tauri::menu::{Menu, MenuItem};
use tauri::tray::{TrayIcon, TrayIconBuilder};
use tauri::{Emitter, Manager};
use tauri_plugin_autostart::{MacosLauncher, ManagerExt as AutostartManagerExt};
use tauri_plugin_notification::NotificationExt;
use tracing_appender::{non_blocking, rolling};
use tracing_subscriber::{fmt, prelude::*, EnvFilter};

use crate::core::auth;
use crate::core::connection::{self, ConnectionStatus};
use crate::core::crypto::LocalCrypto;
use crate::core::ingest::{AuthEmitter, CredsProvider, Ingest, WakeUp};
use crate::core::local_api;
use crate::core::preflight::{self, PreflightInput};
use crate::core::store::LocalStore;
use crate::core::watcher::{self, PauseFlag};

struct AgentState {
    store: Arc<LocalStore>,
    data_dir: std::path::PathBuf,
    local_api_port: u16,
    pause: PauseFlag,
    http: reqwest::Client,
    wake: WakeUp,
}

// TrayRef — небольшой контейнер для tray-иконки и pending-item, чтобы
// фоновый loop мог обновлять заголовок "Pending: N".
struct TrayRef {
    #[allow(dead_code)]
    tray: TrayIcon,
    pending_item: MenuItem<tauri::Wry>,
}

// LogGuard — держит `WorkerGuard` от tracing-appender. Drop сбрасывает
// буфер на диск. Храним в Tauri AppHandle state чтобы жил всю сессию.
#[allow(dead_code)]
struct LogGuard(Option<non_blocking::WorkerGuard>);

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    crate::core::env::load_env_files();

    tauri::Builder::default()
        .plugin(tauri_plugin_autostart::init(
            MacosLauncher::LaunchAgent,
            Some(vec!["--silent"]),
        ))
        .plugin(tauri_plugin_notification::init())
        .invoke_handler(tauri::generate_handler![
            status,
            pending_count,
            check_accessibility,
            permissions_status,
            open_permissions_settings,
            connection_status,
            set_paused,
            is_paused,
            account_info,
            pair_begin,
            pair_poll,
            logout,
            open_logs_folder,
            open_url,
            get_autostart,
            set_autostart,
        ])
        .setup(|app| {
            // Tray — добавлен pending-count item (disabled, обновляется раз в 5с).
            let show = MenuItem::with_id(app, "show", "Open dashboard", true, None::<&str>)?;
            let pending_item =
                MenuItem::with_id(app, "pending", "Pending: …", false, None::<&str>)?;
            let pause = MenuItem::with_id(app, "pause", "Pause tracking", true, None::<&str>)?;
            let settings = MenuItem::with_id(app, "settings", "Settings", true, None::<&str>)?;
            let quit = MenuItem::with_id(app, "quit", "Quit", true, None::<&str>)?;
            let menu = Menu::with_items(app, &[&show, &pending_item, &pause, &settings, &quit])?;
            let tray = TrayIconBuilder::new()
                .menu(&menu)
                .tooltip("Eye of Providence")
                .on_menu_event(|app, event| match event.id.as_ref() {
                    "show" | "settings" => {
                        if let Some(window) = app.get_webview_window("main") {
                            let _ = window.show();
                            let _ = window.set_focus();
                        }
                    }
                    "pause" => {
                        let st = app.state::<AgentState>();
                        let new_state = !st.pause.paused();
                        st.pause.set(new_state);
                        tracing::info!(paused = new_state, "tray pause toggle");
                    }
                    "quit" => app.exit(0),
                    _ => {}
                })
                .build(app)?;
            app.manage(TrayRef { tray, pending_item });

            // Panic hook → отдельный crash-лог в app_data_dir/logs/. Полезно когда
            // process падает до того, как async tracing успел сбросить буфер.
            let crash_dir = app.path().app_data_dir()?.join("logs");
            let _ = std::fs::create_dir_all(&crash_dir);
            std::panic::set_hook(Box::new(move |info| {
                let ts = chrono::Utc::now().format("%Y%m%dT%H%M%S");
                let path = crash_dir.join(format!("crash-{}.log", ts));
                let _ = std::fs::write(&path, format!("{info}\n"));
                eprintln!("agent panicked, see {}", path.display());
            }));

            let data_dir = app.path().app_data_dir()?;
            std::fs::create_dir_all(&data_dir)?;

            // Logging: stdout + rotating file (10MB cap по факту через daily rotation;
            // tracing-appender не умеет size-based, но при daily + удаления >5 файлов
            // суммарно log-grows контролируется).
            let logs_dir = data_dir.join("logs");
            std::fs::create_dir_all(&logs_dir)?;
            let file_appender = rolling::Builder::new()
                .filename_prefix("agent")
                .filename_suffix("log")
                .rotation(rolling::Rotation::DAILY)
                .max_log_files(5)
                .build(&logs_dir)
                .expect("rolling file appender");
            let (file_writer, file_guard) = non_blocking(file_appender);
            let env_filter =
                EnvFilter::try_from_default_env().unwrap_or_else(|_| EnvFilter::new("info"));
            tracing_subscriber::registry()
                .with(env_filter)
                .with(fmt::layer().with_writer(std::io::stdout))
                .with(fmt::layer().with_writer(file_writer).with_ansi(false))
                .init();
            app.manage(LogGuard(Some(file_guard)));

            let db_path = data_dir.join("eop.sqlite");
            tracing::info!(path = %db_path.display(), "opening local store");

            // Encryption key — открываем или генерим в Keychain/DPAPI до того
            // как привязываем SQLite. Если keyring недоступен (например, на
            // headless CI или Linux без secret-service), `?` уронит startup —
            // это правильно: plaintext fallback нарушил бы privacy claim из
            // README §8. Лучше hard fail с понятной ошибкой в логе, чем
            // тихий downgrade.
            let crypto = Arc::new(LocalCrypto::open_or_init().map_err(|e| {
                tracing::error!(error = %e, "failed to init local DB encryption key");
                e
            })?);
            let store = Arc::new(LocalStore::open(&db_path, crypto)?);
            let local_api_port: u16 = std::env::var("EOP_LOCAL_API_PORT")
                .ok()
                .and_then(|v| v.parse().ok())
                .unwrap_or(7373);
            let pause = PauseFlag::default();
            let http = reqwest::Client::builder()
                .timeout(Duration::from_secs(10))
                .build()
                .expect("reqwest client");
            let wake = WakeUp::default();
            app.manage(AgentState {
                store: store.clone(),
                data_dir: data_dir.clone(),
                local_api_port,
                pause: pause.clone(),
                http: http.clone(),
                wake: wake.clone(),
            });

            // Стартовые проверки — в лог; UI опрашивает connection_status.
            let pf_dir = data_dir.clone();
            let startup_http = http.clone();
            tauri::async_runtime::spawn(async move {
                for r in preflight::run(PreflightInput { data_dir: &pf_dir }).await {
                    match r.status {
                        preflight::CheckStatus::Ok => {
                            tracing::info!(check = %r.id, "{}", r.message)
                        }
                        preflight::CheckStatus::Error => {
                            tracing::error!(check = %r.id, "{}", r.message)
                        }
                    }
                }
                let conn = connection::status(connection::ConnectionInput {
                    local_api_port,
                    http: &startup_http,
                })
                .await;
                tracing::info!(
                    backend = ?conn.backend,
                    local_api = ?conn.local_api,
                    port = conn.local_api_port,
                    paired = conn.paired,
                    "connection status"
                );
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
            watcher::spawn(
                store.clone(),
                platform,
                Duration::from_secs(5),
                90,
                pause.clone(),
            );

            // Tray pending refresh — обновляем "Pending: N" каждые 5с.
            let tray_store = store.clone();
            let tray_handle = app.handle().clone();
            tauri::async_runtime::spawn(async move {
                loop {
                    tokio::time::sleep(Duration::from_secs(5)).await;
                    let n = tray_store.pending_count().unwrap_or(0);
                    if let Some(tref) = tray_handle.try_state::<TrayRef>() {
                        let _ = tref.pending_item.set_text(format!("Pending: {n}"));
                    }
                }
            });

            // GC loop — раз в час чистим события старше 7 дней.
            let gc_store = store.clone();
            tauri::async_runtime::spawn(async move {
                let interval = Duration::from_secs(3600);
                let max_age_secs: i64 = 7 * 24 * 3600;
                loop {
                    tokio::time::sleep(interval).await;
                    match gc_store.gc(max_age_secs) {
                        Ok(0) => {}
                        Ok(n) => tracing::info!(deleted = n, "store gc — события старше 7d"),
                        Err(err) => tracing::warn!(error = %err, "store gc failed"),
                    }
                }
            });

            // Ingest pump — single, читает креды из keyring каждый tick.
            // logout/pair бьют wake, и loop сам разруливает.
            let app_handle = app.handle().clone();
            let emitter = AuthEmitter::new(move || {
                let _ = app_handle.emit("auth-required", ());
                let _ = app_handle
                    .notification()
                    .builder()
                    .title("Eye of Providence")
                    .body("Re-pair required: token expired or revoked.")
                    .show();
            });
            let creds: CredsProvider = std::sync::Arc::new(|| {
                let url = auth::get_backend_url().ok();
                let tok = auth::get_token().ok().flatten();
                (url, tok)
            });
            Ingest::new(
                store.clone(),
                creds,
                100,
                Duration::from_secs(15),
                emitter,
                wake.clone(),
            )
            .spawn();

            Ok(())
        })
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}

#[tauri::command]
fn status() -> serde_json::Value {
    // version читается из Cargo.toml (CARGO_PKG_VERSION) — единый источник
    // правды; channel держим выровненным с dashboard sidebar version_label.
    serde_json::json!({
        "status": "ok",
        "version": env!("CARGO_PKG_VERSION"),
        "channel": "alpha",
    })
}

#[tauri::command]
fn pending_count(state: tauri::State<'_, AgentState>) -> Result<i64, String> {
    state.store.pending_count().map_err(|e| e.to_string())
}

#[derive(serde::Serialize)]
struct PermissionsStatus {
    accessibility: bool,
    /// Имя в System Settings → Accessibility (включите именно его).
    app_name: String,
    /// Базовое отслеживание работает и без Accessibility.
    optional: bool,
}

#[tauri::command]
fn permissions_status() -> PermissionsStatus {
    #[cfg(target_os = "macos")]
    {
        return PermissionsStatus {
            accessibility: platform::macos::has_accessibility(),
            app_name: platform::macos::accessibility_app_label(),
            optional: true,
        };
    }
    #[cfg(not(target_os = "macos"))]
    {
        // На Windows/Linux нет macOS Accessibility-разрешения; модуль
        // platform::macos cfg'нут наружу, его функции звать нельзя.
        PermissionsStatus {
            accessibility: true,
            app_name: String::new(),
            optional: true,
        }
    }
}

#[tauri::command]
fn open_permissions_settings() -> Result<(), String> {
    #[cfg(target_os = "macos")]
    {
        open_target(
            "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility",
        )
        .map_err(|e| e.to_string())?;
    }
    #[cfg(target_os = "windows")]
    {
        open_target("ms-settings:privacy").map_err(|e| e.to_string())?;
    }
    #[cfg(target_os = "linux")]
    {
        let _ = std::process::Command::new("xdg-open")
            .arg("https://wiki.archlinux.org/title/XDG_Desktop_Portal")
            .spawn();
    }
    Ok(())
}

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

#[tauri::command]
fn set_paused(paused: bool, state: tauri::State<'_, AgentState>) {
    state.pause.set(paused);
    tracing::info!(paused, "tracking pause via UI");
}

#[tauri::command]
fn is_paused(state: tauri::State<'_, AgentState>) -> bool {
    state.pause.paused()
}

#[tauri::command]
async fn connection_status(
    state: tauri::State<'_, AgentState>,
) -> Result<ConnectionStatus, String> {
    Ok(connection::status(connection::ConnectionInput {
        local_api_port: state.local_api_port,
        http: &state.http,
    })
    .await)
}

#[tauri::command]
fn account_info() -> Result<auth::AccountInfo, String> {
    auth::account_info().map_err(|e| e.to_string())
}

#[tauri::command]
async fn pair_begin(
    state: tauri::State<'_, AgentState>,
) -> Result<auth::PairBeginResponse, String> {
    let backend = auth::get_backend_url().map_err(|e| e.to_string())?;
    auth::pair_begin(&state.http, &backend)
        .await
        .map_err(|e| e.to_string())
}

#[tauri::command]
async fn pair_poll(
    pair_id: String,
    secret: String,
    state: tauri::State<'_, AgentState>,
) -> Result<auth::PollResponse, String> {
    let backend = auth::get_backend_url().map_err(|e| e.to_string())?;
    let resp = auth::pair_poll(&state.http, &backend, &pair_id, &secret)
        .await
        .map_err(|e| e.to_string())?;
    if resp.status == "claimed" {
        if let (Some(token), Some(user_id)) = (resp.token.as_deref(), resp.user_id.as_deref()) {
            auth::persist_pairing(token, user_id).map_err(|e| e.to_string())?;
            state.wake.notify();
        }
    }
    Ok(resp)
}

#[tauri::command]
fn logout(state: tauri::State<'_, AgentState>) -> Result<(), String> {
    auth::logout().map_err(|e| e.to_string())?;
    state.wake.notify();
    Ok(())
}

#[tauri::command]
fn open_logs_folder(state: tauri::State<'_, AgentState>) -> Result<String, String> {
    let logs = state.data_dir.join("logs");
    std::fs::create_dir_all(&logs).map_err(|e| e.to_string())?;
    let logs_str = logs.to_string_lossy().to_string();
    open_target(&logs_str).map_err(|e| e.to_string())?;
    Ok(logs_str)
}

#[tauri::command]
fn open_url(url: String) -> Result<(), String> {
    open_target(&url).map_err(|e| e.to_string())
}

#[tauri::command]
fn get_autostart(app: tauri::AppHandle) -> Result<bool, String> {
    app.autolaunch().is_enabled().map_err(|e| e.to_string())
}

#[tauri::command]
fn set_autostart(app: tauri::AppHandle, enabled: bool) -> Result<(), String> {
    let mgr = app.autolaunch();
    if enabled {
        mgr.enable().map_err(|e| e.to_string())
    } else {
        mgr.disable().map_err(|e| e.to_string())
    }
}

// open_target — кроссплатформенно открывает URL/папку через системные хелперы.
// macOS — `open`, Windows — `cmd /C start`, Linux — `xdg-open`. Принимает
// и file-path, и https-URL — оба варианта окей для shell-команд.
fn open_target(target: &str) -> std::io::Result<()> {
    #[cfg(target_os = "macos")]
    {
        std::process::Command::new("open").arg(target).spawn()?;
    }
    #[cfg(target_os = "windows")]
    {
        std::process::Command::new("cmd")
            .args(["/C", "start", ""])
            .arg(target)
            .spawn()?;
    }
    #[cfg(target_os = "linux")]
    {
        std::process::Command::new("xdg-open").arg(target).spawn()?;
    }
    Ok(())
}

fn rand_byte() -> u8 {
    use std::time::{SystemTime, UNIX_EPOCH};
    let n = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.subsec_nanos())
        .unwrap_or(0);
    (n & 0xff) as u8
}
