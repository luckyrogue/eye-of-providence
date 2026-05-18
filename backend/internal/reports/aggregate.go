package reports

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/reports/domain"
	"github.com/eye-of-providence/backend/internal/store"
)

type NumericContext = domain.NumericContext
type LanguageBreakdown = domain.LanguageBreakdown
type LanguageStat = domain.LanguageStat

func BuildContext(ctx context.Context, st store.EventStore, userID string, period string, from, to time.Time) (*domain.NumericContext, error) {
	events, err := st.ListRecent(ctx, userID, 10_000)
	if err != nil {
		return nil, err
	}

	c := &domain.NumericContext{
		Period:            period,
		From:              from,
		To:                to,
		ByCategoryMS:      map[string]uint64{},
		AICharsByProvider: map[string]uint64{},
		AIMSByChannel:     map[string]uint64{},
		ByLanguageChars:   map[string]domain.LanguageBreakdown{},
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
