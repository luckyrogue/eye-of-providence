// macOS watcher — NSWorkspace для foreground app, CGEventSource для idle.
// Phase 1: только foreground app + idle seconds (без keystroke counts —
// требует Accessibility permission, добавим в Phase 1.5).

use super::PlatformWatcher;

pub struct MacosWatcher;

impl MacosWatcher {
    pub fn new() -> Self {
        Self
    }
}

impl PlatformWatcher for MacosWatcher {
    fn current_app(&self) -> Option<String> {
        frontmost_bundle_id()
    }

    fn idle_seconds(&self) -> u32 {
        cg_idle_seconds()
    }
}

#[cfg(target_os = "macos")]
fn frontmost_bundle_id() -> Option<String> {
    use objc2::rc::autoreleasepool;
    use objc2_app_kit::NSWorkspace;

    autoreleasepool(|_| unsafe {
        let workspace = NSWorkspace::sharedWorkspace();
        let app = workspace.frontmostApplication()?;
        let bid = app.bundleIdentifier()?;
        Some(bid.to_string())
    })
}

#[cfg(not(target_os = "macos"))]
fn frontmost_bundle_id() -> Option<String> {
    None
}

#[cfg(target_os = "macos")]
fn cg_idle_seconds() -> u32 {
    // CGEventSourceSecondsSinceLastEventType — публичный API CoreGraphics.
    // state=0 — combinedSessionState, event_type=u32::MAX — kCGAnyInputEventType.
    #[link(name = "ApplicationServices", kind = "framework")]
    unsafe extern "C" {
        fn CGEventSourceSecondsSinceLastEventType(state: u32, event_type: u32) -> f64;
    }
    const COMBINED_SESSION_STATE: u32 = 0;
    const ANY_INPUT_EVENT: u32 = u32::MAX;
    let secs = unsafe { CGEventSourceSecondsSinceLastEventType(COMBINED_SESSION_STATE, ANY_INPUT_EVENT) };
    if secs.is_finite() && secs >= 0.0 {
        secs as u32
    } else {
        0
    }
}

#[cfg(not(target_os = "macos"))]
fn cg_idle_seconds() -> u32 {
    0
}
