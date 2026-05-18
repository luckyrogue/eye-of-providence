// Стартовые проверки окружения (лог при запуске). Сеть и порты — в connection.rs.

use serde::Serialize;
use std::path::Path;

#[derive(Serialize, Clone)]
#[serde(rename_all = "lowercase")]
pub enum CheckStatus {
    Ok,
    Error,
}

#[derive(Serialize, Clone)]
pub struct CheckResult {
    pub id: String,
    pub label: String,
    pub status: CheckStatus,
    pub message: String,
}

pub struct PreflightInput<'a> {
    pub data_dir: &'a Path,
}

pub async fn run(input: PreflightInput<'_>) -> Vec<CheckResult> {
    vec![check_accessibility(), check_data_dir(input.data_dir)]
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
                message: "Granted.".into(),
            };
        }
        CheckResult {
            id,
            label,
            status: CheckStatus::Ok,
            message: "Не выдан — базовые счётчики работают без него.".into(),
        }
    }
    #[cfg(not(target_os = "macos"))]
    {
        CheckResult {
            id,
            label,
            status: CheckStatus::Ok,
            message: "Не требуется на этой платформе.".into(),
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
            }
        }
        Err(err) => CheckResult {
            id,
            label,
            status: CheckStatus::Error,
            message: format!("Не могу писать в {}: {}", dir.display(), err),
        },
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
}
