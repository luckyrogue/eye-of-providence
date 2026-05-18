package anomalyapp

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/store"
)

type EventStoreAdapter struct {
	Store store.EventStore
}

func (a EventStoreAdapter) ActiveUserIDs(ctx context.Context, since time.Time) ([]string, error) {
	return a.Store.ActiveUserIDs(ctx, since)
}

func (a EventStoreAdapter) DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]TrendPoint, error) {
	pts, err := a.Store.DailyTrend(ctx, userID, since, tz)
	if err != nil {
		return nil, err
	}
	out := make([]TrendPoint, 0, len(pts))
	for _, p := range pts {
		out = append(out, TrendPoint{Date: p.Date, Category: p.Category, MS: p.MS})
	}
	return out, nil
}
