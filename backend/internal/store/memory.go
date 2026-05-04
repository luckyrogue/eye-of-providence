package store

import (
	"context"
	"sort"
	"sync"
	"time"
)

// MemoryStore — in-memory реализация EventStore для dev / тестов.
// В production заменяется на ClickHouse store (Phase 2+).
type MemoryStore struct {
	mu     sync.RWMutex
	events []Event
}

func NewMemory() *MemoryStore {
	return &MemoryStore{events: make([]Event, 0, 1024)}
}

func (s *MemoryStore) Insert(_ context.Context, events []Event) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, events...)
	return nil
}

func (s *MemoryStore) ListRecent(_ context.Context, userID string, limit int) ([]Event, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]Event, 0, limit)
	for i := len(s.events) - 1; i >= 0 && len(out) < limit; i-- {
		if s.events[i].UserID == userID {
			out = append(out, s.events[i])
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].TS.After(out[j].TS) })
	return out, nil
}

func (s *MemoryStore) AggregateByCategory(_ context.Context, userID string, since time.Time) (map[string]uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agg := map[string]uint64{}
	for _, e := range s.events {
		if e.UserID != userID || e.TS.Before(since) {
			continue
		}
		agg[e.Category] += uint64(e.DurationMS)
	}
	return agg, nil
}

func (s *MemoryStore) Heatmap(_ context.Context, userID string, since time.Time) ([]HeatmapCell, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type key struct {
		dow, hour int
		category  string
	}
	agg := map[key]uint64{}
	for _, e := range s.events {
		if e.UserID != userID || e.TS.Before(since) {
			continue
		}
		k := key{int(e.TS.UTC().Weekday()), e.TS.UTC().Hour(), e.Category}
		agg[k] += uint64(e.DurationMS)
	}
	out := make([]HeatmapCell, 0, len(agg))
	for k, ms := range agg {
		out = append(out, HeatmapCell{DayOfWeek: k.dow, Hour: k.hour, Category: k.category, MS: ms})
	}
	return out, nil
}

func (s *MemoryStore) Close() error { return nil }
