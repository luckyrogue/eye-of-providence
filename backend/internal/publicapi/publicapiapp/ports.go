package publicapiapp

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/publicapi/domain"
)

type EventReadStore interface {
	ListRecent(ctx context.Context, userID string, limit int) ([]domain.Event, error)
	AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error)
	LanguageBreakdown(ctx context.Context, userID string, since time.Time) ([]domain.LangCell, error)
	DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]domain.TrendPoint, error)
}
