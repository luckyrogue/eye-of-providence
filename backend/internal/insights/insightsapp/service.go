package insightsapp

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/eye-of-providence/backend/internal/insights/domain"
)

type Service struct {
	events EventReadStore
	ranges RangeAggregator
}

type Deps struct {
	Events EventReadStore
	Ranges RangeAggregator
}

func New(d Deps) *Service {
	return &Service{events: d.Events, ranges: d.Ranges}
}

func (s *Service) Generate(ctx context.Context, userID, tz string, now time.Time) ([]domain.Insight, error) {
	last7 := now.Add(-7 * 24 * time.Hour)
	prev7 := now.Add(-14 * 24 * time.Hour)
	last30 := now.Add(-30 * 24 * time.Hour)

	var (
		aggLast, aggPrev map[string]uint64
		langs            []domain.LangCell
		trend            []domain.TrendPoint
	)
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		v, err := s.events.AggregateByCategory(gctx, userID, last7)
		if err != nil {
			return err
		}
		aggLast = v
		return nil
	})
	g.Go(func() error {
		v, err := s.ranges.AggregateWindow(gctx, userID, prev7, last7)
		if err != nil {
			return err
		}
		aggPrev = v
		return nil
	})
	g.Go(func() error {
		v, err := s.events.LanguageBreakdown(gctx, userID, last30)
		if err != nil {
			return err
		}
		langs = v
		return nil
	})
	g.Go(func() error {
		v, err := s.events.DailyTrend(gctx, userID, last7, tz)
		if err != nil {
			return err
		}
		trend = v
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return domain.Generate(domain.Inputs{
		Last7d: aggLast, Prev7d: aggPrev, Languages: langs, Trend: trend,
	}), nil
}
