// Watcher loop: каждые 5 сек спрашивает PlatformWatcher, пишет focus events в LocalStore.
//
// Что трекаем за tick:
//  - foreground app + idle → focus event с duration_ms (как было).
//  - delta keystrokes / mouse clicks → попадают в `chars_in` фокус-события
//    и meta["mouse_clicks"] (для AI vs manual classification на бэке).
//  - изменение clipboard → отдельное `clipboard_signal` событие с sha256.
//    Backend сравнит хеш с paste-событиями IDE/browser плагинов.

use std::sync::atomic::{AtomicBool, Ordering};
use std::sync::Arc;
use std::time::Duration;

use super::event::{Category, Event};
use super::store::LocalStore;
use crate::platform::PlatformWatcher;

/// PauseFlag — общий флаг для всего агента: tray-кнопка "Pause tracking" поднимает
/// его, watcher и ingest pump читают и приостанавливают работу. Atomic — потому что
/// читается из разных tokio-задач без локов.
#[derive(Clone, Default)]
pub struct PauseFlag(pub Arc<AtomicBool>);

impl PauseFlag {
    pub fn paused(&self) -> bool {
        self.0.load(Ordering::Relaxed)
    }
    pub fn set(&self, v: bool) {
        self.0.store(v, Ordering::Relaxed);
    }
}

pub fn spawn(
    store: Arc<LocalStore>,
    watcher: Arc<dyn PlatformWatcher>,
    poll_interval: Duration,
    idle_threshold_secs: u32,
    pause: PauseFlag,
) -> tauri::async_runtime::JoinHandle<()> {
    tauri::async_runtime::spawn(async move {
        let mut last_app: Option<String> = None;
        let mut accumulated_ms: u32 = 0;
        let mut accumulated_keys: u32 = 0;
        let mut accumulated_clicks: u32 = 0;

        // Baseline'ы CG-counters и clipboard. Берём до первого sleep,
        // чтобы первая итерация считала delta от запуска агента, а не
        // от boot'а (cumulative counter с boot времени иначе даст
        // огромный спайк в первом focus-event'е).
        let mut last_keys: u64 = watcher.keystroke_count();
        let mut last_clicks: u64 = watcher.mouse_click_count();
        let mut last_clipboard: i64 = watcher.clipboard_change_count();

        loop {
            tokio::time::sleep(poll_interval).await;
            if pause.paused() {
                if let Some(app) = last_app.take() {
                    flush_focus(&store, &app, accumulated_ms, accumulated_keys, accumulated_clicks);
                    accumulated_ms = 0;
                    accumulated_keys = 0;
                    accumulated_clicks = 0;
                }
                // Не сбрасываем baseline'ы — после resume снова возьмём
                // правильную delta от текущего момента.
                last_keys = watcher.keystroke_count();
                last_clicks = watcher.mouse_click_count();
                last_clipboard = watcher.clipboard_change_count();
                continue;
            }

            let idle = watcher.idle_seconds();
            let current = watcher.current_app();

            // Считаем delta. Counters могут «откатиться» (после reboot / user
            // logout-logout) — в этом случае берём 0, не пытаемся угадать.
            let keys_now = watcher.keystroke_count();
            let clicks_now = watcher.mouse_click_count();
            let delta_keys = keys_now.saturating_sub(last_keys) as u32;
            let delta_clicks = clicks_now.saturating_sub(last_clicks) as u32;
            last_keys = keys_now;
            last_clicks = clicks_now;

            // Clipboard — только если changeCount изменился. Содержимое читается
            // лишь в этом случае (на macOS 14+ читать каждый tick = OS-spam).
            let clipboard_now = watcher.clipboard_change_count();
            if clipboard_now != last_clipboard {
                last_clipboard = clipboard_now;
                if let Some(digest) = watcher.clipboard_digest() {
                    let app_for_clip = current.clone().unwrap_or_else(|| "__unknown__".into());
                    let ev = Event::clipboard_signal(app_for_clip, digest.sha256_hex, digest.size_bytes);
                    if let Err(err) = store.push(&ev) {
                        tracing::warn!(error = %err, "store.push clipboard failed");
                    }
                }
            }

            // idle > threshold → flush текущий focus, переключаемся в idle
            if idle >= idle_threshold_secs {
                if let Some(app) = last_app.take() {
                    flush_focus(&store, &app, accumulated_ms, accumulated_keys, accumulated_clicks);
                    accumulated_ms = 0;
                    accumulated_keys = 0;
                    accumulated_clicks = 0;
                }
                let idle_event =
                    Event::os_focus("__idle__", Category::Idle, poll_interval.as_millis() as u32);
                let _ = store.push(&idle_event);
                continue;
            }

            match (&last_app, &current) {
                (Some(prev), Some(now)) if prev == now => {
                    accumulated_ms += poll_interval.as_millis() as u32;
                    accumulated_keys = accumulated_keys.saturating_add(delta_keys);
                    accumulated_clicks = accumulated_clicks.saturating_add(delta_clicks);
                }
                (Some(prev), Some(_)) => {
                    flush_focus(&store, prev, accumulated_ms, accumulated_keys, accumulated_clicks);
                    accumulated_ms = poll_interval.as_millis() as u32;
                    accumulated_keys = delta_keys;
                    accumulated_clicks = delta_clicks;
                    last_app = current;
                }
                (Some(prev), None) => {
                    flush_focus(&store, prev, accumulated_ms, accumulated_keys, accumulated_clicks);
                    accumulated_ms = 0;
                    accumulated_keys = 0;
                    accumulated_clicks = 0;
                    last_app = None;
                }
                (None, Some(_)) => {
                    accumulated_ms = poll_interval.as_millis() as u32;
                    accumulated_keys = delta_keys;
                    accumulated_clicks = delta_clicks;
                    last_app = current;
                }
                (None, None) => {}
            }
        }
    })
}

fn flush_focus(store: &LocalStore, app: &str, ms: u32, keys: u32, clicks: u32) {
    if ms == 0 {
        return;
    }
    let category = classify_app(app);
    let mut event = Event::os_focus(app, category, ms);
    event.chars_in = keys;
    if clicks > 0 {
        event.meta.insert("mouse_clicks".to_string(), clicks.to_string());
    }
    if let Err(err) = store.push(&event) {
        tracing::warn!(error = %err, "store.push failed");
    }
}

/// Очень грубая классификация по bundle id. Phase 3 заменяется на attribution-pipeline.
fn classify_app(app_bundle: &str) -> Category {
    let b = app_bundle.to_ascii_lowercase();
    if b.contains("vscode")
        || b.contains("jetbrains")
        || b.contains("cursor")
        || b.contains("xcode")
    {
        Category::Manual
    } else if b.contains("safari")
        || b.contains("chrome")
        || b.contains("firefox")
        || b.contains("arc")
    {
        Category::Other // browser события приходят отдельно от extension
    } else {
        Category::Other
    }
}
