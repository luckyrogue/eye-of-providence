// Загрузка `.env` при старте агента (dev: ищем вверх от cwd; prod: EOP_ENV_FILE).

/// Не перезаписывает уже заданные в процессе переменные.
pub fn load_env_files() {
    if let Ok(path) = std::env::var("EOP_ENV_FILE") {
        if !path.trim().is_empty() {
            let _ = dotenvy::from_filename_override(&path);
            return;
        }
    }

    if let Ok(cwd) = std::env::current_dir() {
        let mut dir: &std::path::Path = cwd.as_path();
        for _ in 0..8 {
            let candidate = dir.join(".env");
            if candidate.is_file() {
                let _ = dotenvy::from_path(&candidate);
                return;
            }
            dir = match dir.parent() {
                Some(p) => p,
                None => break,
            };
        }
    }
}

pub fn non_empty(key: &str) -> Option<String> {
    std::env::var(key)
        .ok()
        .map(|v| v.trim().to_string())
        .filter(|v| !v.is_empty())
}
