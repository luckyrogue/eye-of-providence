// Platform-specific watchers. Тонкие модули — каждый ~500 строк.

#[cfg(target_os = "macos")]
pub mod macos;

#[cfg(target_os = "windows")]
pub mod windows;

pub trait PlatformWatcher: Send + Sync {
    fn current_app(&self) -> Option<String>;
    fn idle_seconds(&self) -> u32;
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
    fn current_app(&self) -> Option<String> { None }
    fn idle_seconds(&self) -> u32 { 0 }
}
