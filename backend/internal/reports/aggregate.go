package reports

import (
	"context"
	"sort"
	"time"

	"github.com/eye-of-providence/backend/internal/store"
)

// NumericContext — что мы шлём в Gemini.
// Жёсткое правило: только числа и метки (категория, провайдер, язык).
// Никакого контента файлов, промптов, заголовков, меты с url/title.
type NumericContext struct {
	Period            string                       `json:"period"`             // weekly_2026_W18
	From              time.Time                    `json:"from"`
	To                time.Time                    `json:"to"`
	TotalActiveMS     uint64                       `json:"total_active_ms"`
	IdleMS            uint64                       `json:"idle_ms"`
	ByCategoryMS      map[string]uint64            `json:"by_category_ms"`
	AICharsByProvider map[string]uint64            `json:"ai_chars_by_provider"`
	AIMSByChannel     map[string]uint64            `json:"ai_ms_by_channel"`
	ByLanguageChars   map[string]LanguageBreakdown `json:"by_language_chars"`
	EventsTotal       int                          `json:"events_total"`
}

type LanguageBreakdown struct {
	ManualChars uint64 `json:"manual_chars"`
	AIChars     uint64 `json:"ai_chars"`
	TotalMS     uint64 `json:"total_ms"`
}

// BuildContext — собирает агрегаты из EventStore за указанный период.
func BuildContext(ctx context.Context, st store.EventStore, userID string, period string, from, to time.Time) (*NumericContext, error) {
	// Используем универсальный путь: ListRecent по большому limit, фильтруем по from/to.
	// В Phase 4 заменяется на правильные ClickHouse aggregations.
	events, err := st.ListRecent(ctx, userID, 10_000)
	if err != nil {
		return nil, err
	}

	c := &NumericContext{
		Period:            period,
		From:              from,
		To:                to,
		ByCategoryMS:      map[string]uint64{},
		AICharsByProvider: map[string]uint64{},
		AIMSByChannel:     map[string]uint64{},
		ByLanguageChars:   map[string]LanguageBreakdown{},
	}

	for _, e := range events {
		if e.TS.Before(from) || e.TS.After(to) {
			continue
		}
		c.EventsTotal++
		c.ByCategoryMS[e.Category] += uint64(e.DurationMS)
		if e.Category == "idle" {
			c.IdleMS += uint64(e.DurationMS)
		} else {
			c.TotalActiveMS += uint64(e.DurationMS)
		}
		if e.Category == "ai" {
			if e.AIProvider != "" {
				c.AICharsByProvider[e.AIProvider] += uint64(e.CharsIn)
			}
			if e.AIChannel != "" {
				c.AIMSByChannel[e.AIChannel] += uint64(e.DurationMS)
			}
		}
		if e.FileLang != "" {
			lb := c.ByLanguageChars[e.FileLang]
			if e.Category == "ai" {
				lb.AIChars += uint64(e.CharsIn)
			} else if e.Category == "manual" || e.Category == "refactor" {
				lb.ManualChars += uint64(e.CharsIn)
			}
			lb.TotalMS += uint64(e.DurationMS)
			c.ByLanguageChars[e.FileLang] = lb
		}
	}
	return c, nil
}

// TopLanguages — для удобства промпта: top-N языков по сумме chars (manual+ai).
func (c *NumericContext) TopLanguages(n int) []LanguageStat {
	stats := make([]LanguageStat, 0, len(c.ByLanguageChars))
	for lang, b := range c.ByLanguageChars {
		stats = append(stats, LanguageStat{Lang: lang, Manual: b.ManualChars, AI: b.AIChars, MS: b.TotalMS})
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Manual+stats[i].AI > stats[j].Manual+stats[j].AI
	})
	if len(stats) > n {
		stats = stats[:n]
	}
	return stats
}

type LanguageStat struct {
	Lang   string
	Manual uint64
	AI     uint64
	MS     uint64
}
