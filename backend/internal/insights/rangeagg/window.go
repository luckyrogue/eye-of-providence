package rangeagg

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
)

// Store — минимум для окна [since, until) через diff двух AggregateByCategory.
type Store interface {
	AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error)
}

// AggregateWindow — agg(since) − agg(until) (см. insights handler).
func AggregateWindow(ctx context.Context, st Store, userID string, since, until time.Time) (map[string]uint64, error) {
	var full, tail map[string]uint64
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		v, err := st.AggregateByCategory(gctx, userID, since)
		if err != nil {
			return err
		}
		full = v
		return nil
	})
	g.Go(func() error {
		v, err := st.AggregateByCategory(gctx, userID, until)
		if err != nil {
			return err
		}
		tail = v
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	out := make(map[string]uint64, len(full))
	for k, v := range full {
		t := tail[k]
		if v >= t {
			out[k] = v - t
		}
	}
	return out, nil
}
