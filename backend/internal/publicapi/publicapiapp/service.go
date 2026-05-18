package publicapiapp

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/publicapi/domain"
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

func (s *Service) Summary(ctx context.Context, userID string, since time.Time) (map[string]uint64, error) {
	if s.events == nil {
		return map[string]uint64{}, nil
	}
	return s.events.AggregateByCategory(ctx, userID, since)
}

func (s *Service) Languages(ctx context.Context, userID string, since time.Time) ([]domain.LangCell, error) {
	if s.events == nil {
		return nil, nil
	}
	return s.events.LanguageBreakdown(ctx, userID, since)
}

func (s *Service) Trend(ctx context.Context, userID string, since time.Time, tz string) ([]domain.TrendPoint, error) {
	if s.events == nil {
		return nil, nil
	}
	return s.events.DailyTrend(ctx, userID, since, tz)
}

func WindowFromDays(days int) (int, time.Time) {
	since := time.Now().UTC().Add(-time.Duration(days) * 24 * time.Hour)
	return days, since
}
