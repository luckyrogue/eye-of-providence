package store

import (
	"context"
	"time"
)

// Event — внутренняя модель события (без proto зависимости пока в Phase 1).
// Соответствует proto/event.proto, но используем плоские строки для категорий
// чтобы было удобно сериализовать в JSON и хранить в любом store.
type Event struct {
	TS           time.Time `json:"ts"`
	UserID       string    `json:"user_id"`
	DeviceID     string    `json:"device_id"`
	SessionID    string    `json:"session_id"`
	AppBundle    string    `json:"app_bundle"`
	Category     string    `json:"category"` // idle | manual | ai | reading | refactor | other
	Source       string    `json:"source"`   // os | browser | ide | cli
	AIProvider   string    `json:"ai_provider,omitempty"`
	AIChannel    string    `json:"ai_channel,omitempty"`
	ProjectID    string    `json:"project_id,omitempty"`
	FileLang     string    `json:"file_lang,omitempty"`
	DurationMS   uint32    `json:"duration_ms"`
	CharsIn      uint32    `json:"chars_in"`
	LinesAdded   uint32    `json:"lines_added"`
	LinesRemoved uint32    `json:"lines_removed"`
}

type EventStore interface {
	Insert(ctx context.Context, events []Event) error
	ListRecent(ctx context.Context, userID string, limit int) ([]Event, error)
	AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error)
	Close() error
}
