package store

import (
	"context"
	"time"
)

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
	// Meta — небольшой набор безопасных полей расширения (нет контента,
	// см. docs/api/openapi.yaml Event.meta). Используется агентом для
	// clipboard_sha256 + clipboard_bytes (фингерпринт без содержимого).
	// Сериализуется как JSON-объект в ClickHouse `meta` String-колонку.
	Meta map[string]string `json:"meta,omitempty"`
}

type EventStore interface {
	Insert(ctx context.Context, events []Event) error
	ListRecent(ctx context.Context, userID string, limit int) ([]Event, error)
	AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error)

	// AggregateProvenance — суммы focus_ms по code-provenance категориям
	// (`typed`, `pasted_ai`, `pasted_other`, `ai_inline`, `ai_agent`,
	// `refactor`, `unknown`) из таблицы `attribution_events`, которую пишет
	// attribution worker. Это НЕ то же, что AggregateByCategory: там сырые
	// категории событий (`manual`/`ai`/`idle`/…), здесь — результат
	// классификации provenance.
	AggregateProvenance(ctx context.Context, userID string, since time.Time) (map[string]uint64, error)

	AggregateByCategoryBulk(ctx context.Context, userIDs []string, since time.Time) (map[string]map[string]uint64, error)
	Heatmap(ctx context.Context, userID string, since time.Time, tz string) ([]HeatmapCell, error)
	LanguageBreakdown(ctx context.Context, userID string, since time.Time) ([]LangCell, error)
	ActiveUserIDs(ctx context.Context, since time.Time) ([]string, error)
	DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]TrendPoint, error)
	Close() error
}

type TrendPoint struct {
	Date     string `json:"date"`
	Category string `json:"category"`
	Chars    uint64 `json:"chars"`
	MS       uint64 `json:"ms"`
}

type HeatmapCell struct {
	DayOfWeek int    `json:"dow"`
	Hour      int    `json:"hour"`
	Category  string `json:"category"`
	MS        uint64 `json:"ms"`
}

type LangCell struct {
	Lang     string `json:"lang"`
	Category string `json:"category"`
	Chars    uint64 `json:"chars"`
	MS       uint64 `json:"ms"`
}
