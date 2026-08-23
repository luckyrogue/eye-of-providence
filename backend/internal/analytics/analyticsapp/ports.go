package analyticsapp

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/analytics/domain"
)

// EventReadStore — narrow read port for analytics BC.
type EventReadStore interface {
	ListRecent(ctx context.Context, userID string, limit int) ([]domain.Event, error)
	LanguageBreakdown(ctx context.Context, userID string, since time.Time) ([]domain.LangCell, error)
	DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]domain.TrendPoint, error)
	Heatmap(ctx context.Context, userID string, since time.Time, tz string) ([]domain.HeatmapCell, error)
	AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error)
	AggregateProvenance(ctx context.Context, userID string, since time.Time) (map[string]uint64, error)
}
