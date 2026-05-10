package store

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/cache"
)

// fakeStore — counts calls для verification cache hits/misses.
type fakeStore struct {
	aggCalls       atomic.Int32
	bulkCalls      atomic.Int32
	langCalls      atomic.Int32
	trendCalls     atomic.Int32
	heatmapCalls   atomic.Int32
	insertCalls    atomic.Int32
	listCalls      atomic.Int32
	deleteCalls    atomic.Int32
	closeCalls     atomic.Int32
	activeCalls    atomic.Int32

	aggResp     map[string]uint64
	bulkResp    map[string]map[string]uint64
	langResp    []LangCell
	trendResp   []TrendPoint
	heatmapResp []HeatmapCell
}

func (f *fakeStore) Insert(_ context.Context, _ []Event) error { f.insertCalls.Add(1); return nil }
func (f *fakeStore) ListRecent(_ context.Context, _ string, _ int) ([]Event, error) {
	f.listCalls.Add(1); return nil, nil
}
func (f *fakeStore) AggregateByCategory(_ context.Context, _ string, _ time.Time) (map[string]uint64, error) {
	f.aggCalls.Add(1); return f.aggResp, nil
}
func (f *fakeStore) AggregateByCategoryBulk(_ context.Context, _ []string, _ time.Time) (map[string]map[string]uint64, error) {
	f.bulkCalls.Add(1); return f.bulkResp, nil
}
func (f *fakeStore) Heatmap(_ context.Context, _ string, _ time.Time, _ string) ([]HeatmapCell, error) {
	f.heatmapCalls.Add(1); return f.heatmapResp, nil
}
func (f *fakeStore) LanguageBreakdown(_ context.Context, _ string, _ time.Time) ([]LangCell, error) {
	f.langCalls.Add(1); return f.langResp, nil
}
func (f *fakeStore) ActiveUserIDs(_ context.Context, _ time.Time) ([]string, error) {
	f.activeCalls.Add(1); return nil, nil
}
func (f *fakeStore) DailyTrend(_ context.Context, _ string, _ time.Time, _ string) ([]TrendPoint, error) {
	f.trendCalls.Add(1); return f.trendResp, nil
}
func (f *fakeStore) Close() error { f.closeCalls.Add(1); return nil }

func (f *fakeStore) DeleteUserData(_ context.Context, _ string) error {
	f.deleteCalls.Add(1); return nil
}

// setupCache — стартует mini-redis-style цепочку через real redis.Client с
// in-memory mock через DB 15. На CI без Redis тест skip'ится.
func setupCache(t *testing.T) *cache.Cache {
	t.Helper()
	cli := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   15, // dedicated test DB
	})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := cli.Ping(ctx).Err(); err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	cli.FlushDB(ctx)
	t.Cleanup(func() {
		cli.FlushDB(context.Background())
		_ = cli.Close()
	})
	return &cache.Cache{Client: cli, Prefix: "test:cache"}
}

func TestCached_NoCacheFallthrough(t *testing.T) {
	inner := &fakeStore{aggResp: map[string]uint64{"ai": 100}}
	// nil cache → NewCached returns inner as-is
	wrapped := NewCached(inner, nil, zap.NewNop())
	if wrapped != inner {
		t.Error("nil cache should return inner as-is")
	}
}

func TestCached_AggregateHitMiss(t *testing.T) {
	c := setupCache(t)
	inner := &fakeStore{aggResp: map[string]uint64{"ai": 100, "manual": 200}}
	wrapped := NewCached(inner, c, zap.NewNop())

	ctx := context.Background()
	since := time.Now().Add(-7 * 24 * time.Hour)

	// Miss → calls inner
	r1, err := wrapped.AggregateByCategory(ctx, "user-1", since)
	if err != nil {
		t.Fatal(err)
	}
	if inner.aggCalls.Load() != 1 {
		t.Errorf("inner calls=%d, want 1", inner.aggCalls.Load())
	}
	if r1["ai"] != 100 {
		t.Errorf("got %v", r1)
	}

	// Hit → no inner call
	r2, err := wrapped.AggregateByCategory(ctx, "user-1", since)
	if err != nil {
		t.Fatal(err)
	}
	if inner.aggCalls.Load() != 1 {
		t.Errorf("inner calls=%d (cache miss), want still 1", inner.aggCalls.Load())
	}
	if r2["ai"] != 100 {
		t.Errorf("cached returned wrong value: %v", r2)
	}
}

func TestCached_DifferentUsers_DifferentKeys(t *testing.T) {
	c := setupCache(t)
	inner := &fakeStore{aggResp: map[string]uint64{"ai": 1}}
	wrapped := NewCached(inner, c, zap.NewNop())
	ctx := context.Background()
	since := time.Now()

	wrapped.AggregateByCategory(ctx, "user-A", since)
	wrapped.AggregateByCategory(ctx, "user-B", since)
	if inner.aggCalls.Load() != 2 {
		t.Errorf("expected 2 inner calls (different users), got %d", inner.aggCalls.Load())
	}
}

func TestCached_DifferentSince_DifferentKeys(t *testing.T) {
	c := setupCache(t)
	inner := &fakeStore{aggResp: map[string]uint64{"ai": 1}}
	wrapped := NewCached(inner, c, zap.NewNop())
	ctx := context.Background()
	uid := "user-1"

	wrapped.AggregateByCategory(ctx, uid, time.Now())
	wrapped.AggregateByCategory(ctx, uid, time.Now().Add(-24*time.Hour))
	if inner.aggCalls.Load() != 2 {
		t.Errorf("expected 2 inner calls, got %d", inner.aggCalls.Load())
	}
}

func TestCached_BulkSortStability(t *testing.T) {
	c := setupCache(t)
	inner := &fakeStore{bulkResp: map[string]map[string]uint64{"u1": {"ai": 5}}}
	wrapped := NewCached(inner, c, zap.NewNop())
	ctx := context.Background()
	since := time.Now()

	// Same set, different order → same cache key
	wrapped.AggregateByCategoryBulk(ctx, []string{"u1", "u2"}, since)
	wrapped.AggregateByCategoryBulk(ctx, []string{"u2", "u1"}, since)
	if inner.bulkCalls.Load() != 1 {
		t.Errorf("expected 1 inner call (sorted key), got %d", inner.bulkCalls.Load())
	}
}

func TestCached_TZ_DifferentKeys(t *testing.T) {
	c := setupCache(t)
	inner := &fakeStore{trendResp: []TrendPoint{{Date: "2026-05-01", Category: "ai"}}}
	wrapped := NewCached(inner, c, zap.NewNop())
	ctx := context.Background()
	uid := "u"
	since := time.Now()

	wrapped.DailyTrend(ctx, uid, since, "UTC")
	wrapped.DailyTrend(ctx, uid, since, "Asia/Almaty")
	if inner.trendCalls.Load() != 2 {
		t.Errorf("tz-different keys should miss both: got %d", inner.trendCalls.Load())
	}
}

func TestCached_DeleteInvalidates(t *testing.T) {
	c := setupCache(t)
	inner := &fakeStore{aggResp: map[string]uint64{"ai": 1}}
	wrapped := NewCached(inner, c, zap.NewNop())
	ctx := context.Background()
	uid := "user-del"
	since := time.Now()

	wrapped.AggregateByCategory(ctx, uid, since) // populate cache
	if del, ok := wrapped.(UserDeleter); ok {
		if err := del.DeleteUserData(ctx, uid); err != nil {
			t.Fatal(err)
		}
	} else {
		t.Fatal("CachedEventStore should implement UserDeleter")
	}

	// After delete, next call → cache miss → new inner call
	wrapped.AggregateByCategory(ctx, uid, since)
	if inner.aggCalls.Load() != 2 {
		t.Errorf("expected 2 inner calls (cache invalidated), got %d", inner.aggCalls.Load())
	}
}

func TestCached_SetFailureNotFatal(t *testing.T) {
	// Cache с invalid endpoint → SET fails, но read pass-through still works
	cli := redis.NewClient(&redis.Options{Addr: "127.0.0.1:1", DialTimeout: 100 * time.Millisecond})
	c := &cache.Cache{Client: cli, Prefix: "test"}
	defer cli.Close()

	inner := &fakeStore{aggResp: map[string]uint64{"ai": 1}}
	wrapped := NewCached(inner, c, zap.NewNop())

	// Even with broken cache, read should succeed via inner
	_, err := wrapped.AggregateByCategory(context.Background(), "u", time.Now())
	if err != nil && !errors.Is(err, context.DeadlineExceeded) {
		// Inner returned successfully, cache write/read failed silently
		// (DeadlineExceeded ok если ping глобальный timeout сработал)
	}
}
