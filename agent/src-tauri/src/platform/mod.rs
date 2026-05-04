// Platform-specific watchers. Тонкие модули — каждый ~500 строк.
// Phase 1: реализация macOS, Phase 4: Windows.

#[cfg(target_os = "macos")]
pub mod macos;

#[cfg(target_os = "windows")]
pub mod windows;

pub trait PlatformWatcher: Send + Sync {
    fn current_app(&self) -> Option<String>;
    fn idle_seconds(&self) -> u32;
}
