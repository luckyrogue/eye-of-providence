// Windows watcher — GetForegroundWindow + GetLastInputInfo.
// Не требует никаких прав admin / accessibility, работает из коробки.

use super::PlatformWatcher;

pub struct WindowsWatcher;

impl WindowsWatcher {
    pub fn new() -> Self {
        Self
    }
}

impl PlatformWatcher for WindowsWatcher {
    fn current_app(&self) -> Option<String> {
        foreground_process_name()
    }

    fn idle_seconds(&self) -> u32 {
        last_input_seconds()
    }
}

#[cfg(target_os = "windows")]
fn foreground_process_name() -> Option<String> {
    use windows::Win32::Foundation::{CloseHandle, MAX_PATH};
    use windows::Win32::System::Threading::{
        OpenProcess, QueryFullProcessImageNameW, PROCESS_QUERY_LIMITED_INFORMATION,
    };
    use windows::Win32::UI::WindowsAndMessaging::{GetForegroundWindow, GetWindowThreadProcessId};

    unsafe {
        let hwnd = GetForegroundWindow();
        if hwnd.0.is_null() {
            return None;
        }
        let mut pid: u32 = 0;
        GetWindowThreadProcessId(hwnd, Some(&mut pid));
        if pid == 0 {
            return None;
        }
        let handle = OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION, false, pid).ok()?;

        let mut buf = [0u16; MAX_PATH as usize];
        let mut size = buf.len() as u32;
        let res = QueryFullProcessImageNameW(
            handle,
            Default::default(),
            windows::core::PWSTR(buf.as_mut_ptr()),
            &mut size,
        );
        let _ = CloseHandle(handle);
        if res.is_err() {
            return None;
        }
        let path = String::from_utf16_lossy(&buf[..size as usize]);
        // Возвращаем только имя файла (как app_bundle proxy), без полного пути —
        // чтобы избежать раскрытия username и других PII.
        let exe = std::path::Path::new(&path)
            .file_name()
            .and_then(|s| s.to_str())
            .map(String::from);
        exe
    }
}

#[cfg(not(target_os = "windows"))]
fn foreground_process_name() -> Option<String> {
    None
}

#[cfg(target_os = "windows")]
fn last_input_seconds() -> u32 {
    use windows::Win32::System::SystemInformation::GetTickCount;
    use windows::Win32::UI::Input::KeyboardAndMouse::{GetLastInputInfo, LASTINPUTINFO};

    unsafe {
        let mut info = LASTINPUTINFO {
            cbSize: std::mem::size_of::<LASTINPUTINFO>() as u32,
            dwTime: 0,
        };
        // GetLastInputInfo возвращает `BOOL`, в `windows` 0.58 у BOOL нет
        // `is_err()` — это `windows::core::BOOL`, обёртка над i32. Проверяем
        // через `.as_bool()` (true ↔ success). Реверс — если false, значит
        // системный вызов провалился (теоретически возможно при low-memory),
        // отдаём 0 = «активен» чтобы не дропать события.
        if !GetLastInputInfo(&mut info).as_bool() {
            return 0;
        }
        let now = GetTickCount();
        let elapsed_ms = now.saturating_sub(info.dwTime);
        elapsed_ms / 1000
    }
}

#[cfg(not(target_os = "windows"))]
fn last_input_seconds() -> u32 {
    0
}
