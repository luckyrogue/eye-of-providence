// Preflight checks: проверяем окружение при старте агента и при открытии Onboarding.
// Каждая проверка возвращает CheckResult с status (ok/warn/error) и hint что делать.

use serde::Serialize;
use std::net::TcpListener;
use std::path::Path;
use std::time::Duration;

#[derive(Serialize, Clone)]
#[serde(rename_all = "lowercase")]
pub enum CheckStatus {
    Ok,
    Warn,
    Error,
}

#[derive(Serialize, Clone)]
pub struct CheckResult {
    pub id: String,
    pub label: String,
    pub status: CheckStatus,
    pub message: String,
    /// Если есть — UI рендерит кнопку с этим URL.
    #[serde(skip_serializing_if = "Option::is_none")]
    pub action_url: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub action_label: Option<String>,
}

pub struct PreflightInput<'a> {
    pub data_dir: &'a Path,
    pub local_api_port: u16,
    pub check_local_api_port: bool,
    pub backend_url: Option<&'a str>,
    pub bearer_token: Option<&'a str>,
}

pub async fn run(input: PreflightInput<'_>) -> Vec<CheckResult> {
    let mut out = vec![check_accessibility(), check_data_dir(input.data_dir)];
    out.push(if input.check_local_api_port {
        check_local_api_port(input.local_api_port)
    } else {
        CheckResult {
            id: "local_api_port".to_string(),
            label: format!("Local API port {} available", input.local_api_port),
            status: CheckStatus::Ok,
            message: "Skipped port availability check (local API may already be running).".into(),
            action_url: None,
            action_label: None,
        }
    });
    out.push(check_backend(input.backend_url, input.bearer_token).await);
    out
}

fn check_accessibility() -> CheckResult {
    let id = "accessibility".to_string();
    let label = "macOS Accessibility permission".to_string();

    #[cfg(target_os = "macos")]
    {
        if crate::platform::macos::has_accessibility() {
            return CheckResult {
                id,
                label,
                status: CheckStatus::Ok,
                message: "Granted. Доступны расширенные сигналы для attribution.".into(),
                action_url: None,
                action_label: None,
            };
        }
        // Базовые keystroke/mouse counts работают БЕЗ Accessibility — через
        // публичный CGEventSourceCounterForEventType. Accessibility теперь
        // нужен только для будущих фич (focused text-field detection и т.п.),
        // поэтому статус «info», а не «warn».
        CheckResult {
            id,
            label,
            status: CheckStatus::Ok,
            message: "Не выдан. Базовые keystroke counts работают; Accessibility опционален для будущих фич.".into(),
            action_url: Some(
                "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility"
                    .into(),
            ),
            action_label: Some("Open System Settings".into()),
        }
    }
    #[cfg(not(target_os = "macos"))]
    {
        CheckResult {
            id,
            label,
            status: CheckStatus::Ok,
            message: "Не требуется на этой платформе.".into(),
            action_url: None,
            action_label: None,
        }
    }
}

fn check_data_dir(dir: &Path) -> CheckResult {
    let id = "data_dir".to_string();
    let label = "Local data directory writable".to_string();
    let probe = dir.join(".eop-write-probe");
    match std::fs::write(&probe, b"ok") {
        Ok(_) => {
            let _ = std::fs::remove_file(&probe);
            CheckResult {
                id,
                label,
                status: CheckStatus::Ok,
                message: format!("OK: {}", dir.display()),
                action_url: None,
                action_label: None,
            }
        }
        Err(err) => CheckResult {
            id,
            label,
            status: CheckStatus::Error,
            message: format!("Не могу писать в {}: {}", dir.display(), err),
            action_url: None,
            action_label: None,
        },
    }
}

fn check_local_api_port(port: u16) -> CheckResult {
    let id = "local_api_port".to_string();
    let label = format!("Local API port {} available", port);
    match TcpListener::bind(("127.0.0.1", port)) {
        Ok(listener) => {
            drop(listener);
            CheckResult {
                id,
                label,
                status: CheckStatus::Ok,
                message: "Свободен. Browser extension и IDE plugin подключатся.".into(),
                action_url: None,
                action_label: None,
            }
        }
        Err(_) => CheckResult {
            id,
            label,
            status: CheckStatus::Warn,
            message: format!(
                "Порт {} занят. Закрой процесс или смени EOP_LOCAL_API_PORT.",
                port
            ),
            action_url: None,
            action_label: None,
        },
    }
}

async fn check_backend(url: Option<&str>, token: Option<&str>) -> CheckResult {
    let id = "backend".to_string();
    let label = "Backend reachable".to_string();
    let Some(url) = url else {
        return CheckResult {
            id,
            label,
            status: CheckStatus::Warn,
            message: "EOP_BACKEND_URL не задан — local-only режим (события в SQLite).".into(),
            action_url: Some("https://github.com/luckyrogue/eye-of-providence#deploy".into()),
            action_label: Some("Docs".into()),
        };
    };
    let healthz = format!("{}/healthz", url.trim_end_matches('/'));
    let client = match reqwest::Client::builder()
        .timeout(Duration::from_secs(3))
        .build()
    {
        Ok(c) => c,
        Err(_) => return error_check(id, label, "не удалось создать HTTP client".into()),
    };
    match client.get(&healthz).send().await {
        Ok(r) if r.status().is_success() && token.is_some() => CheckResult {
            id,
            label,
            status: CheckStatus::Ok,
            message: format!("OK: {} (token задан)", url),
            action_url: None,
            action_label: None,
        },
        Ok(r) if r.status().is_success() => CheckResult {
            id,
            label,
            status: CheckStatus::Warn,
            message: format!("{} отвечает, но EOP_BEARER_TOKEN не задан.", url),
            action_url: None,
            action_label: None,
        },
        Ok(r) => error_check(id, label, format!("{} вернул {}", url, r.status())),
        Err(err) => error_check(id, label, format!("{} недоступен: {}", url, err)),
    }
}

fn error_check(id: String, label: String, message: String) -> CheckResult {
    CheckResult {
        id,
        label,
        status: CheckStatus::Error,
        message,
        action_url: None,
        action_label: None,
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use tempfile::tempdir;

    #[test]
    fn data_dir_writable_ok() {
        let dir = tempdir().unwrap();
        let r = check_data_dir(dir.path());
        assert!(matches!(r.status, CheckStatus::Ok));
        assert_eq!(r.id, "data_dir");
    }

    #[test]
    fn data_dir_missing_errors() {
        let r = check_data_dir(Path::new("/definitely/does/not/exist/eop"));
        assert!(matches!(r.status, CheckStatus::Error));
    }

    #[tokio::test]
    async fn backend_missing_url_warns() {
        let r = check_backend(None, None).await;
        assert!(matches!(r.status, CheckStatus::Warn));
        assert_eq!(r.id, "backend");
    }
}
