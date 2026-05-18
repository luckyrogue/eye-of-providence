// Event types — mirror proto/event.proto.
// JSON-сериализация для отправки в backend (Phase 1).

use std::collections::BTreeMap;

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
    /// Безопасные расширения без контента: clipboard_sha256/clipboard_bytes,
    /// mouse_clicks и т.п. См. proto/event.proto Event.meta. BTreeMap чтобы
    /// JSON-сериализация была детерминированной (помогает в тестах и при
    /// дедупе хешей событий на бэкенде).
    #[serde(default, skip_serializing_if = "BTreeMap::is_empty")]
    pub meta: BTreeMap<String, String>,
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
            meta: BTreeMap::new(),
        }
    }

    /// Clipboard-fingerprint событие: НЕ содержит содержимого, только sha256
    /// + размер.
    ///
    /// Backend сравнит хеш с paste-событиями из IDE/browser плагинов
    /// чтобы атрибутировать вставку (typed vs pasted-AI vs pasted-other).
    pub fn clipboard_signal(
        app_bundle: impl Into<String>,
        sha256_hex: String,
        size_bytes: usize,
    ) -> Self {
        let mut meta = BTreeMap::new();
        meta.insert("clipboard_sha256".to_string(), sha256_hex);
        meta.insert("clipboard_bytes".to_string(), size_bytes.to_string());
        Self {
            ts: Utc::now(),
            user_id: String::new(),
            device_id: String::new(),
            session_id: String::new(),
            app_bundle: app_bundle.into(),
            category: Category::Other,
            source: Source::Os,
            ai_provider: None,
            ai_channel: None,
            project_id: None,
            file_lang: None,
            duration_ms: 0,
            chars_in: 0,
            lines_added: 0,
            lines_removed: 0,
            meta,
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn clipboard_signal_meta_is_deterministic() {
        let ev = Event::clipboard_signal("com.google.Chrome", "abc123".into(), 42);
        let json = serde_json::to_string(&ev).unwrap();
        // BTreeMap гарантирует алфавитный порядок ключей → clipboard_bytes < clipboard_sha256.
        assert!(json.contains(r#""meta":{"clipboard_bytes":"42","clipboard_sha256":"abc123"}"#));
        assert_eq!(ev.category, Category::Other);
        assert_eq!(ev.source, Source::Os);
        assert_eq!(ev.duration_ms, 0);
    }

    #[test]
    fn focus_event_serializes_without_meta_when_empty() {
        let ev = Event::os_focus("com.microsoft.VSCode", Category::Manual, 5000);
        let json = serde_json::to_string(&ev).unwrap();
        // Пустой meta → skip_serializing_if прячет поле (контракт с backend
        // omitempty Meta map).
        assert!(!json.contains("\"meta\""));
    }
}
