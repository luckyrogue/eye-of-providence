// macOS watcher — NSWorkspace для foreground app, CoreGraphics для idle и
// keystroke/mouse counters, NSPasteboard для clipboard fingerprint.
//
// Permissions: keystroke/mouse counts через `CGEventSourceCounterForEventType`
// — публичный CG API, НЕ требует Accessibility (в отличие от CGEventTap).
// Clipboard — NSPasteboard, тоже без спец-разрешений; на macOS 14+ Sequoia
// сам OS показывает уведомление «X accessed Clipboard» при чтении data —
// поэтому читаем содержимое (для sha256) только когда changeCount изменился.

use super::{ClipboardDigest, PlatformWatcher};

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

    fn keystroke_count(&self) -> u64 {
        cg_event_counter(K_CG_EVENT_KEY_DOWN)
    }

    fn mouse_click_count(&self) -> u64 {
        cg_event_counter(K_CG_EVENT_LEFT_MOUSE_DOWN) + cg_event_counter(K_CG_EVENT_RIGHT_MOUSE_DOWN)
    }

    fn clipboard_change_count(&self) -> i64 {
        pasteboard_change_count()
    }

    fn clipboard_digest(&self) -> Option<ClipboardDigest> {
        pasteboard_digest()
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

/// Имя процесса в System Settings → Privacy → Accessibility (может отличаться от
/// `eop-agent` в dev и `Eye of Providence` в .app bundle).
#[cfg(target_os = "macos")]
pub fn accessibility_app_label() -> String {
    if let Ok(exe) = std::env::current_exe() {
        let path = exe.to_string_lossy();
        if path.contains("Eye of Providence") {
            return "Eye of Providence".into();
        }
        if let Some(stem) = exe.file_stem() {
            return stem.to_string_lossy().into_owned();
        }
    }
    "Eye of Providence".into()
}

/// Проверяет Accessibility для **текущего** процесса (`AXIsProcessTrusted`).
/// После включения в System Settings нужен перезапуск агента.
#[cfg(target_os = "macos")]
pub fn has_accessibility() -> bool {
    #[link(name = "ApplicationServices", kind = "framework")]
    unsafe extern "C" {
        fn AXIsProcessTrusted() -> u8;
    }
    unsafe { AXIsProcessTrusted() != 0 }
}

#[cfg(not(target_os = "macos"))]
pub fn accessibility_app_label() -> String {
    "Eye of Providence".into()
}

#[cfg(not(target_os = "macos"))]
pub fn has_accessibility() -> bool {
    true
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
    let secs =
        unsafe { CGEventSourceSecondsSinceLastEventType(COMBINED_SESSION_STATE, ANY_INPUT_EVENT) };
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

// CGEventType константы — из <CoreGraphics/CGEventTypes.h>.
const K_CG_EVENT_LEFT_MOUSE_DOWN: u32 = 1;
const K_CG_EVENT_RIGHT_MOUSE_DOWN: u32 = 3;
const K_CG_EVENT_KEY_DOWN: u32 = 10;

#[cfg(target_os = "macos")]
fn cg_event_counter(event_type: u32) -> u64 {
    // CGEventSourceCounterForEventType — публичный CG API. Возвращает
    // монотонно растущий счётчик событий с момента user login. Не
    // требует Accessibility permission (в отличие от CGEventTap).
    // state=0 — combinedSessionState, считаем по всем источникам.
    #[link(name = "ApplicationServices", kind = "framework")]
    unsafe extern "C" {
        fn CGEventSourceCounterForEventType(state: u32, event_type: u32) -> u32;
    }
    const COMBINED_SESSION_STATE: u32 = 0;
    unsafe { CGEventSourceCounterForEventType(COMBINED_SESSION_STATE, event_type) as u64 }
}

#[cfg(not(target_os = "macos"))]
fn cg_event_counter(_event_type: u32) -> u64 {
    0
}

#[cfg(target_os = "macos")]
fn pasteboard_change_count() -> i64 {
    use objc2::rc::autoreleasepool;
    use objc2_app_kit::NSPasteboard;
    autoreleasepool(|_| unsafe {
        let pb = NSPasteboard::generalPasteboard();
        pb.changeCount() as i64
    })
}

#[cfg(not(target_os = "macos"))]
fn pasteboard_change_count() -> i64 {
    0
}

#[cfg(target_os = "macos")]
fn pasteboard_digest() -> Option<ClipboardDigest> {
    use objc2::rc::autoreleasepool;
    use objc2_app_kit::NSPasteboard;
    use objc2_foundation::NSString;
    use sha2::{Digest, Sha256};

    // public.utf8-plain-text — стандартный UTI для текста на macOS;
    // покрывает paste из browser/IDE. Если clipboard содержит rich
    // content (image/file), fingerprint пропускаем — для AI-attribution
    // важны только текстовые вставки.
    autoreleasepool(|_| unsafe {
        let pb = NSPasteboard::generalPasteboard();
        let utf8_type = NSString::from_str("public.utf8-plain-text");
        let data = pb.dataForType(&utf8_type)?;
        let len = data.len();
        if len == 0 {
            return None;
        }
        let bytes_ptr = data.bytes().as_ptr();
        let slice = std::slice::from_raw_parts(bytes_ptr, len);
        let mut hasher = Sha256::new();
        hasher.update(slice);
        let digest = hasher.finalize();
        Some(ClipboardDigest {
            sha256_hex: hex::encode(digest),
            size_bytes: len,
        })
    })
}

#[cfg(not(target_os = "macos"))]
fn pasteboard_digest() -> Option<ClipboardDigest> {
    None
}
