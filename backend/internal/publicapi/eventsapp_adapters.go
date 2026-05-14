package publicapi

import (
	"context"

	"github.com/eye-of-providence/backend/internal/publicapi/eventsapp"
	"github.com/eye-of-providence/backend/internal/store"
)

type eventsListReader struct{ st store.EventStore }

func (e eventsListReader) ListRecent(ctx context.Context, userID string, limit int) ([]store.Event, error) {
	return e.st.ListRecent(ctx, userID, limit)
}

func newEventsApp(st store.EventStore) *eventsapp.Service {
	return eventsapp.New(eventsListReader{st: st})
}
