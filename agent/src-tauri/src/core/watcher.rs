// Watcher loop: каждые 5 сек спрашивает PlatformWatcher, пишет focus events в LocalStore.

use std::sync::Arc;
use std::time::Duration;

use super::event::{Category, Event};
use super::store::LocalStore;
use crate::platform::PlatformWatcher;

pub fn spawn(
    store: Arc<LocalStore>,
    watcher: Arc<dyn PlatformWatcher>,
    poll_interval: Duration,
    idle_threshold_secs: u32,
) -> tokio::task::JoinHandle<()> {
    tokio::spawn(async move {
        let mut last_app: Option<String> = None;
        let mut accumulated_ms: u32 = 0;

        loop {
            tokio::time::sleep(poll_interval).await;

            let idle = watcher.idle_seconds();
            let current = watcher.current_app();

            // idle > threshold → flush текущий focus, переключаемся в idle
            if idle >= idle_threshold_secs {
                if let Some(app) = last_app.take() {
                    flush_focus(&store, &app, accumulated_ms);
                    accumulated_ms = 0;
                }
                let idle_event = Event::os_focus("__idle__", Category::Idle, poll_interval.as_millis() as u32);
                let _ = store.push(&idle_event);
                continue;
            }

            match (&last_app, &current) {
                (Some(prev), Some(now)) if prev == now => {
                    accumulated_ms += poll_interval.as_millis() as u32;
                }
                (Some(prev), Some(_)) => {
                    flush_focus(&store, prev, accumulated_ms);
                    accumulated_ms = poll_interval.as_millis() as u32;
                    last_app = current;
                }
                (Some(prev), None) => {
                    flush_focus(&store, prev, accumulated_ms);
                    accumulated_ms = 0;
                    last_app = None;
                }
                (None, Some(_)) => {
                    accumulated_ms = poll_interval.as_millis() as u32;
                    last_app = current;
                }
                (None, None) => {}
            }
        }
    })
}

fn flush_focus(store: &LocalStore, app: &str, ms: u32) {
    if ms == 0 {
        return;
    }
    let category = classify_app(app);
    let event = Event::os_focus(app, category, ms);
    if let Err(err) = store.push(&event) {
        tracing::warn!(error = %err, "store.push failed");
    }
}

/// Очень грубая классификация по bundle id. Phase 3 заменяется на attribution-pipeline.
fn classify_app(app_bundle: &str) -> Category {
    let b = app_bundle.to_ascii_lowercase();
    if b.contains("vscode") || b.contains("jetbrains") || b.contains("cursor") || b.contains("xcode") {
        Category::Manual
    } else if b.contains("safari") || b.contains("chrome") || b.contains("firefox") || b.contains("arc") {
        Category::Other // browser события приходят отдельно от extension
    } else {
        Category::Other
    }
}
