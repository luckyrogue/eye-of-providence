package domain

import (
	"sort"
	"time"
)

type NumericContext struct {
	Period            string                       `json:"period"`
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

type LanguageStat struct {
	Lang   string
	Manual uint64
	AI     uint64
	MS     uint64
}

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
