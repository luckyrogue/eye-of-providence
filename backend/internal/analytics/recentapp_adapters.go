package analytics

import (
	"context"

	"github.com/eye-of-providence/backend/internal/analytics/recentapp"
	"github.com/eye-of-providence/backend/internal/store"
)

type analyticsRecentReader struct{ st store.EventStore }

func (a analyticsRecentReader) ListRecent(ctx context.Context, userID string, limit int) ([]store.Event, error) {
	return a.st.ListRecent(ctx, userID, limit)
}

func newRecentEventsApp(st store.EventStore) *recentapp.Service {
	return recentapp.New(analyticsRecentReader{st: st})
}
