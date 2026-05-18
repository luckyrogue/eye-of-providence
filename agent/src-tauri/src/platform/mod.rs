// Platform-specific watchers. Тонкие модули — каждый ~500 строк.

#[cfg(target_os = "macos")]
pub mod macos;

#[cfg(target_os = "windows")]
pub mod windows;

/// Снимок clipboard'а — sha256 содержимого + размер. Содержимое сюда НЕ
/// попадает (см. README §8 privacy claim). Используется только для
/// AI-attribution: сравнение хеша с paste-событиями из IDE/extension.
pub struct ClipboardDigest {
    pub sha256_hex: String,
    pub size_bytes: usize,
}

pub trait PlatformWatcher: Send + Sync {
    fn current_app(&self) -> Option<String>;
    fn idle_seconds(&self) -> u32;

    /// Cumulative keystroke count с момента user login. Watcher loop
    /// вычитает baseline предыдущего тика и получает delta. 0 = платформа
    /// не поддерживает или ещё не было ввода.
    fn keystroke_count(&self) -> u64 {
        0
    }

    /// Cumulative mouse-click count (left+right) с момента user login.
    fn mouse_click_count(&self) -> u64 {
        0
    }

    /// Текущий `NSPasteboard.changeCount` (или Windows clipboard sequence).
    /// Монотонно растёт при каждом изменении clipboard. Дёшево — не
    /// читает содержимое, поэтому опрашиваем каждый tick.
    fn clipboard_change_count(&self) -> i64 {
        0
    }

    /// Читает clipboard и возвращает sha256 + размер. Вызывать только
    /// когда `clipboard_change_count` изменился — на macOS 14+ это
    /// триггерит системное уведомление «X accessed Clipboard», так что
    /// частить нельзя.
    fn clipboard_digest(&self) -> Option<ClipboardDigest> {
        None
    }
}

#[cfg(target_os = "macos")]
pub fn build() -> std::sync::Arc<dyn PlatformWatcher> {
    std::sync::Arc::new(macos::MacosWatcher::new())
}

#[cfg(target_os = "windows")]
pub fn build() -> std::sync::Arc<dyn PlatformWatcher> {
    std::sync::Arc::new(windows::WindowsWatcher::new())
}

#[cfg(not(any(target_os = "macos", target_os = "windows")))]
pub fn build() -> std::sync::Arc<dyn PlatformWatcher> {
    std::sync::Arc::new(NoopWatcher)
}

#[cfg(not(any(target_os = "macos", target_os = "windows")))]
struct NoopWatcher;

#[cfg(not(any(target_os = "macos", target_os = "windows")))]
impl PlatformWatcher for NoopWatcher {
    fn current_app(&self) -> Option<String> {
        None
    }
    fn idle_seconds(&self) -> u32 {
        0
    }
}
