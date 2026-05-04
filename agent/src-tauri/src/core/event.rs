// Event types — mirror proto/event.proto.
// JSON-сериализация для отправки в backend (Phase 1).

use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
pub enum Source {
    Os,
    Browser,
    Ide,
    Cli,
}

#[derive(Debug, Clone, Copy, Serialize, Deserialize, PartialEq, Eq)]
#[serde(rename_all = "lowercase")]
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
    pub ts: DateTime<Utc>,
    pub user_id: String,
    pub device_id: String,
    pub session_id: String,
    pub app_bundle: String,
    pub category: Category,
    pub source: Source,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ai_provider: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub ai_channel: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub project_id: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub file_lang: Option<String>,
    pub duration_ms: u32,
    pub chars_in: u32,
    pub lines_added: u32,
    pub lines_removed: u32,
}

impl Event {
    pub fn os_focus(app_bundle: impl Into<String>, category: Category, duration_ms: u32) -> Self {
        Self {
            ts: Utc::now(),
            user_id: String::new(),
            device_id: String::new(),
            session_id: String::new(),
            app_bundle: app_bundle.into(),
            category,
            source: Source::Os,
            ai_provider: None,
            ai_channel: None,
            project_id: None,
            file_lang: None,
            duration_ms,
            chars_in: 0,
            lines_added: 0,
            lines_removed: 0,
        }
    }
}
