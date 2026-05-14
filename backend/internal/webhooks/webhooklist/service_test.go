package webhooklist_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/eye-of-providence/backend/internal/webhooks/webhooklist"
)

type fakeR struct {
	rows []webhooklist.Row
}

func (f fakeR) ListByUser(ctx context.Context, userID uuid.UUID) ([]webhooklist.Row, error) {
	return f.rows, nil
}

func TestList(t *testing.T) {
	now := time.Now().UTC()
	r := webhooklist.Row{ID: uuid.New(), URL: "https://x", Events: []string{"commit.ingested"}, Format: "raw", Active: true, CreatedAt: now}
	s := webhooklist.New(fakeR{rows: []webhooklist.Row{r}})
	got, err := s.List(context.Background(), uuid.New())
	if err != nil || len(got) != 1 {
		t.Fatal(err, got)
	}
}
