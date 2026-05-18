package domain

import "time"

type Event struct {
	TS           time.Time         `json:"ts"`
	UserID       string            `json:"user_id"`
	DeviceID     string            `json:"device_id"`
	SessionID    string            `json:"session_id"`
	AppBundle    string            `json:"app_bundle"`
	Category     string            `json:"category"`
	Source       string            `json:"source"`
	AIProvider   string            `json:"ai_provider,omitempty"`
	AIChannel    string            `json:"ai_channel,omitempty"`
	ProjectID    string            `json:"project_id,omitempty"`
	FileLang     string            `json:"file_lang,omitempty"`
	DurationMS   uint32            `json:"duration_ms"`
	CharsIn      uint32            `json:"chars_in"`
	LinesAdded   uint32            `json:"lines_added"`
	LinesRemoved uint32            `json:"lines_removed"`
	Meta         map[string]string `json:"meta,omitempty"`
}

func ValidEvent(e Event) bool {
	if e.AppBundle == "" || e.Source == "" || e.Category == "" {
		return false
	}
	switch e.Source {
	case "os", "browser", "ide", "cli":
	default:
		return false
	}
	switch e.Category {
	case "idle", "manual", "ai", "reading", "refactor", "other":
	default:
		return false
	}
	return e.DurationMS <= 24*60*60*1000
}
