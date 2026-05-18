package ingest

import (
	"context"

	"github.com/eye-of-providence/backend/internal/ingest/domain"
	"github.com/eye-of-providence/backend/internal/ingest/ingestapp"
	"github.com/eye-of-providence/backend/internal/store"
)

type eventWriterAdapter struct{ st store.EventStore }

func (a eventWriterAdapter) Insert(ctx context.Context, events []domain.Event) error {
	rows := make([]store.Event, len(events))
	for i := range events {
		rows[i] = mapToStore(events[i])
	}
	return a.st.Insert(ctx, rows)
}

func mapToStore(e domain.Event) store.Event {
	return store.Event{
		TS: e.TS, UserID: e.UserID, DeviceID: e.DeviceID, SessionID: e.SessionID,
		AppBundle: e.AppBundle, Category: e.Category, Source: e.Source,
		AIProvider: e.AIProvider, AIChannel: e.AIChannel, ProjectID: e.ProjectID,
		FileLang: e.FileLang, DurationMS: e.DurationMS, CharsIn: e.CharsIn,
		LinesAdded: e.LinesAdded, LinesRemoved: e.LinesRemoved, Meta: e.Meta,
	}
}

func mapFromStore(e store.Event) domain.Event {
	return domain.Event{
		TS: e.TS, UserID: e.UserID, DeviceID: e.DeviceID, SessionID: e.SessionID,
		AppBundle: e.AppBundle, Category: e.Category, Source: e.Source,
		AIProvider: e.AIProvider, AIChannel: e.AIChannel, ProjectID: e.ProjectID,
		FileLang: e.FileLang, DurationMS: e.DurationMS, CharsIn: e.CharsIn,
		LinesAdded: e.LinesAdded, LinesRemoved: e.LinesRemoved, Meta: e.Meta,
	}
}

func newIngestApp(st store.EventStore) *ingestapp.Service {
	return ingestapp.New(ingestapp.Deps{Writer: eventWriterAdapter{st: st}})
}
