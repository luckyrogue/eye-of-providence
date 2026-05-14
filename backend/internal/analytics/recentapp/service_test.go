package recentapp_test

import (
	"context"
	"testing"

	"github.com/eye-of-providence/backend/internal/analytics/recentapp"
	"github.com/eye-of-providence/backend/internal/store"
)

type mem struct{}

func (mem) ListRecent(ctx context.Context, userID string, limit int) ([]store.Event, error) {
	return []store.Event{{UserID: userID}}, nil
}

func TestListRecent(t *testing.T) {
	s := recentapp.New(mem{})
	got, err := s.ListRecent(context.Background(), "u", 5)
	if err != nil || len(got) != 1 {
		t.Fatal(err, got)
	}
}
