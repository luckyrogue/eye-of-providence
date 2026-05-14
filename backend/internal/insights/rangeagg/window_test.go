package rangeagg_test

import (
	"context"
	"testing"
	"time"

	"github.com/eye-of-providence/backend/internal/insights/rangeagg"
)

type fakeStore struct {
	t1, t2 time.Time
}

func (f fakeStore) AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error) {
	switch {
	case since.Equal(f.t1):
		return map[string]uint64{"a": 100}, nil
	case since.Equal(f.t2):
		return map[string]uint64{"a": 30}, nil
	default:
		return map[string]uint64{}, nil
	}
}

func TestAggregateWindow(t *testing.T) {
	t1 := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2020, 1, 8, 0, 0, 0, 0, time.UTC)
	st := fakeStore{t1: t1, t2: t2}
	got, err := rangeagg.AggregateWindow(context.Background(), st, "u", t1, t2)
	if err != nil || got["a"] != 70 {
		t.Fatalf("%v err=%v", got, err)
	}
}
