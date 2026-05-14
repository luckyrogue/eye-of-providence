package insights

import (
	"context"
	"time"

	"github.com/eye-of-providence/backend/internal/insights/rangeagg"
	"github.com/eye-of-providence/backend/internal/store"
)

type rangeAggStore struct{ st store.EventStore }

func (r rangeAggStore) AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error) {
	return r.st.AggregateByCategory(ctx, userID, since)
}

func aggregateRangeCtx(ctx context.Context, st store.EventStore, userID string, since, until time.Time) (map[string]uint64, error) {
	return rangeagg.AggregateWindow(ctx, rangeAggStore{st: st}, userID, since, until)
}
