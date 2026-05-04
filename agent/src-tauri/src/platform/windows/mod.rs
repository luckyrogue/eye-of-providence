// Windows watcher — SetWinEventHook, GetForegroundWindow, GetLastInputInfo.
// Skeleton: реальная реализация в Phase 4.

use super::PlatformWatcher;

pub struct WindowsWatcher;

impl WindowsWatcher {
    pub fn new() -> Self {
        Self
    }
}

impl PlatformWatcher for WindowsWatcher {
    fn current_app(&self) -> Option<String> {
        // TODO Phase 4: GetForegroundWindow + QueryFullProcessImageName
        None
    }

    fn idle_seconds(&self) -> u32 {
        // TODO Phase 4: GetLastInputInfo
        0
    }
}
