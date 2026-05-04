// Event types — соответствуют proto/event.proto.
// В Phase 0 используем POD-структуры; в Phase 1 заменим на сгенерированные prost-ом.

use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Source {
    Os,
    Browser,
    Ide,
    Cli,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum Category {
    Idle,
    Manual,
    Ai,
    Reading,
    Refactor,
    Other,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Event {
    pub ts_ms: u64,
    pub user_id: String,
    pub device_id: String,
    pub session_id: String,
    pub app_bundle: String,
    pub category: Category,
    pub source: Source,
    pub ai_provider: Option<String>,
    pub project_id: Option<String>,
    pub file_lang: Option<String>,
    pub duration_ms: u32,
    pub chars_in: u32,
    pub lines_added: u32,
    pub lines_removed: u32,
}
