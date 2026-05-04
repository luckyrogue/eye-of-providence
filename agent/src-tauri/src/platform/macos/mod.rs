// macOS watcher — NSWorkspace, CGEventTap, NSPasteboard.
// Skeleton: реальная реализация в Phase 1.

use super::PlatformWatcher;

pub struct MacosWatcher;

impl MacosWatcher {
    pub fn new() -> Self {
        Self
    }
}

impl PlatformWatcher for MacosWatcher {
    fn current_app(&self) -> Option<String> {
        // TODO Phase 1: NSWorkspace.shared.frontmostApplication.bundleIdentifier
        None
    }

    fn idle_seconds(&self) -> u32 {
        // TODO Phase 1: CGEventSourceSecondsSinceLastEventType
        0
    }
}
