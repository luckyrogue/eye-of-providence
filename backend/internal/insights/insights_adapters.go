package insights

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/insights/domain"
	"github.com/eye-of-providence/backend/internal/insights/insightsapp"
	"github.com/eye-of-providence/backend/internal/insights/rangeagg"
	"github.com/eye-of-providence/backend/internal/store"
)

type eventReadAdapter struct{ st store.EventStore }

func (a eventReadAdapter) AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error) {
	return a.st.AggregateByCategory(ctx, userID, since)
}

func (a eventReadAdapter) LanguageBreakdown(ctx context.Context, userID string, since time.Time) ([]domain.LangCell, error) {
	cells, err := a.st.LanguageBreakdown(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	out := make([]domain.LangCell, len(cells))
	for i := range cells {
		out[i] = domain.LangCell(cells[i])
	}
	return out, nil
}

func (a eventReadAdapter) DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]domain.TrendPoint, error) {
	points, err := a.st.DailyTrend(ctx, userID, since, tz)
	if err != nil {
		return nil, err
	}
	out := make([]domain.TrendPoint, len(points))
	for i := range points {
		out[i] = domain.TrendPoint(points[i])
	}
	return out, nil
}

type rangeAggStore struct{ st store.EventStore }

func (r rangeAggStore) AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error) {
	return r.st.AggregateByCategory(ctx, userID, since)
}

type rangeAggAdapter struct{ st store.EventStore }

func (a rangeAggAdapter) AggregateWindow(ctx context.Context, userID string, since, until time.Time) (map[string]uint64, error) {
	return rangeagg.AggregateWindow(ctx, rangeAggStore{st: a.st}, userID, since, until)
}

func newInsightsApp(st store.EventStore) *insightsapp.Service {
	return insightsapp.New(insightsapp.Deps{
		Events: eventReadAdapter{st: st},
		Ranges: rangeAggAdapter{st: st},
	})
}
