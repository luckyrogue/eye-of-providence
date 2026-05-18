package analyticsapp

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/analytics/domain"
)

type Service struct {
	events EventReadStore
}

type Deps struct {
	Events EventReadStore
}

func New(d Deps) *Service {
	return &Service{events: d.Events}
}

func (s *Service) ListRecent(ctx context.Context, userID string, limit int) ([]domain.Event, error) {
	if s.events == nil {
		return []domain.Event{}, nil
	}
	return s.events.ListRecent(ctx, userID, limit)
}

func (s *Service) LanguageBreakdown(ctx context.Context, userID string, w domain.DaysWindow) ([]domain.LangCell, error) {
	if s.events == nil {
		return nil, nil
	}
	return s.events.LanguageBreakdown(ctx, userID, w.Since)
}

func (s *Service) DailyTrend(ctx context.Context, userID string, q domain.TrendQuery) ([]domain.TrendPoint, error) {
	if s.events == nil {
		return nil, nil
	}
	return s.events.DailyTrend(ctx, userID, q.Since, q.TZ)
}

func (s *Service) Heatmap(ctx context.Context, userID string, w domain.DaysWindow, tz string) ([]domain.HeatmapCell, error) {
	if s.events == nil {
		return nil, nil
	}
	return s.events.Heatmap(ctx, userID, w.Since, tz)
}

func (s *Service) Categories(ctx context.Context, userID string, w domain.DaysWindow) (map[string]uint64, error) {
	if s.events == nil {
		return map[string]uint64{}, nil
	}
	return s.events.AggregateByCategory(ctx, userID, w.Since)
}

// WindowFromDays builds a UTC lookback window (days clamped by caller).
func WindowFromDays(days int) domain.DaysWindow {
	since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
	return domain.DaysWindow{Days: days, Since: since}
}

// TrendWindowFromDays truncates since to UTC midnight for daily buckets.
func TrendWindowFromDays(days int) domain.TrendQuery {
	w := WindowFromDays(days)
	w.Since = w.Since.Truncate(24 * time.Hour)
	return domain.TrendQuery{DaysWindow: w, TZ: "UTC"}
}
