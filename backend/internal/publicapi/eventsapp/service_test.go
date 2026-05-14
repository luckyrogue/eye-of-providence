package eventsapp_test

import (
	"context"
	"testing"

	"github.com/eye-of-providence/backend/internal/publicapi/eventsapp"
	"github.com/eye-of-providence/backend/internal/store"
)

type mem struct{}

func (mem) ListRecent(ctx context.Context, userID string, limit int) ([]store.Event, error) {
	return []store.Event{{UserID: userID}}, nil
}

func TestListRecent(t *testing.T) {
	s := eventsapp.New(mem{})
	got, err := s.ListRecent(context.Background(), "u1", 10)
	if err != nil || len(got) != 1 {
		t.Fatal(err, got)
	}
}
