package store

import (
	"context"
	"sort"
	"sync"
	"time"
)

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

// AggregateProvenance для in-memory fallback всегда пуст: provenance
// вычисляет attribution worker, а он пишет в ClickHouse-таблицу
// `attribution_events`, которой в этом сторе не существует. Дублировать здесь
// classify() из internal/attribution нельзя — две копии правил разъедутся.
// Практическое следствие: без ClickHouse донат «Code provenance» пуст.
func (s *MemoryStore) AggregateProvenance(_ context.Context, _ string, _ time.Time) (map[string]uint64, error) {
	return map[string]uint64{}, nil
}

func (s *MemoryStore) AggregateByCategoryBulk(_ context.Context, userIDs []string, since time.Time) (map[string]map[string]uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	want := make(map[string]struct{}, len(userIDs))
	for _, u := range userIDs {
		want[u] = struct{}{}
	}
	out := map[string]map[string]uint64{}
	for _, e := range s.events {
		if e.TS.Before(since) {
			continue
		}
		if _, ok := want[e.UserID]; !ok {
			continue
		}
		if out[e.UserID] == nil {
			out[e.UserID] = map[string]uint64{}
		}
		out[e.UserID][e.Category] += uint64(e.DurationMS)
	}
	return out, nil
}

func (s *MemoryStore) Heatmap(_ context.Context, userID string, since time.Time, tz string) ([]HeatmapCell, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	loc := loadLocation(tz)
	type key struct {
		dow, hour int
		category  string
	}
	agg := map[key]uint64{}
	for _, e := range s.events {
		if e.UserID != userID || e.TS.Before(since) {
			continue
		}
		t := e.TS.In(loc)
		k := key{int(t.Weekday()), t.Hour(), e.Category}
		agg[k] += uint64(e.DurationMS)
	}
	out := make([]HeatmapCell, 0, len(agg))
	for k, ms := range agg {
		out = append(out, HeatmapCell{DayOfWeek: k.dow, Hour: k.hour, Category: k.category, MS: ms})
	}
	return out, nil
}

func (s *MemoryStore) LanguageBreakdown(_ context.Context, userID string, since time.Time) ([]LangCell, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	type key struct{ lang, category string }
	agg := map[key]LangCell{}
	for _, e := range s.events {
		if e.UserID != userID || e.TS.Before(since) || e.FileLang == "" {
			continue
		}
		k := key{e.FileLang, e.Category}
		v := agg[k]
		v.Lang = e.FileLang
		v.Category = e.Category
		v.Chars += uint64(e.CharsIn)
		v.MS += uint64(e.DurationMS)
		agg[k] = v
	}
	out := make([]LangCell, 0, len(agg))
	for _, v := range agg {
		out = append(out, v)
	}
	return out, nil
}

func (s *MemoryStore) ActiveUserIDs(_ context.Context, since time.Time) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := map[string]struct{}{}
	for _, e := range s.events {
		if e.TS.Before(since) {
			continue
		}
		seen[e.UserID] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for u := range seen {
		out = append(out, u)
	}
	return out, nil
}

func (s *MemoryStore) DailyTrend(_ context.Context, userID string, since time.Time, tz string) ([]TrendPoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	loc := loadLocation(tz)
	type key struct{ date, category string }
	agg := map[key]TrendPoint{}
	for _, e := range s.events {
		if e.UserID != userID || e.TS.Before(since) {
			continue
		}
		date := e.TS.In(loc).Format("2006-01-02")
		k := key{date, e.Category}
		v := agg[k]
		v.Date = date
		v.Category = e.Category
		v.Chars += uint64(e.CharsIn)
		v.MS += uint64(e.DurationMS)
		agg[k] = v
	}
	out := make([]TrendPoint, 0, len(agg))
	for _, v := range agg {
		out = append(out, v)
	}
	return out, nil
}

func loadLocation(tz string) *time.Location {
	if tz == "" {
		return time.UTC
	}
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc
	}
	return time.UTC
}

func (s *MemoryStore) Close() error { return nil }
