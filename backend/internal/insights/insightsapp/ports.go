package insightsapp

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/insights/domain"
)

type EventReadStore interface {
	AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error)
	LanguageBreakdown(ctx context.Context, userID string, since time.Time) ([]domain.LangCell, error)
	DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]domain.TrendPoint, error)
}

type RangeAggregator interface {
	AggregateWindow(ctx context.Context, userID string, since, until time.Time) (map[string]uint64, error)
}
