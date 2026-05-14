package pushlist_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/push/pushlist"
)

type fakeR struct{}

func (fakeR) ListByUser(ctx context.Context, userID uuid.UUID) ([]pushlist.SubscriptionRow, error) {
	return []pushlist.SubscriptionRow{{ID: uuid.New(), Endpoint: "https://ep", CreatedAt: "2020-01-01T00:00:00Z"}}, nil
}

func TestList(t *testing.T) {
	s := pushlist.New(fakeR{})
	got, err := s.List(context.Background(), uuid.New())
	if err != nil || len(got) != 1 {
		t.Fatal(err, got)
	}
}
