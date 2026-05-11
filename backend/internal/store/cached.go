package store

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/eye-of-providence/backend/internal/cache"
)

// CachedEventStore — decorator поверх EventStore. Read-heavy aggregations
// идут через Redis cache (TTL 5-10 мин). Mutating ops (Insert) проходят сквозь
// напрямую — eventual consistency приемлема для dashboard'а.
//
// Methods для cache (most-expensive aggregations):
//
//	AggregateByCategory       — TTL 10m
//	AggregateByCategoryBulk   — TTL 5m (multi-user, скорее меняется)
//	LanguageBreakdown         — TTL 10m
//	DailyTrend                — TTL 5m
//	Heatmap                   — TTL 10m
//
// Skipped (always-fresh OR low value):
//
//	ListRecent                — events меняются constantly
//	ActiveUserIDs             — admin-only, infrequent
//	Insert / DeleteUserData   — mutating, no point
type CachedEventStore struct {
	Inner  EventStore
	Cache  *cache.Cache
	Logger *zap.Logger
}

// NewCached — wraps store с cache. Если c == nil — возвращает store as-is.
func NewCached(inner EventStore, c *cache.Cache, logger *zap.Logger) EventStore {
	if c == nil || c.Client == nil {
		return inner
	}
	return &CachedEventStore{Inner: inner, Cache: c, Logger: logger}
}

const (
	ttlAgg     = 10 * time.Minute
	ttlBulk    = 5 * time.Minute
	ttlLang    = 10 * time.Minute
	ttlTrend   = 5 * time.Minute
	ttlHeatmap = 10 * time.Minute
)

func (s *CachedEventStore) Insert(ctx context.Context, events []Event) error {
	return s.Inner.Insert(ctx, events)
}

func (s *CachedEventStore) ListRecent(ctx context.Context, userID string, limit int) ([]Event, error) {
	return s.Inner.ListRecent(ctx, userID, limit)
}

func (s *CachedEventStore) ActiveUserIDs(ctx context.Context, since time.Time) ([]string, error) {
	return s.Inner.ActiveUserIDs(ctx, since)
}

func (s *CachedEventStore) Close() error {
	return s.Inner.Close()
}

// DeleteUserData — pass-through если inner store implements UserDeleter.
// Также инвалидируем cached entries за этого user'а через scan + delete
// (best-effort; если scan падает, TTL очистит eventually).
func (s *CachedEventStore) DeleteUserData(ctx context.Context, userID string) error {
	d, ok := s.Inner.(UserDeleter)
	if !ok {
		return nil
	}
	if err := d.DeleteUserData(ctx, userID); err != nil {
		return err
	}
	s.invalidateUser(ctx, userID)
	return nil
}

// invalidateUser — best-effort удаление всех cache keys для user'а. Match по
// suffix ":{userID}:" (мы пишем prefix:UID:since в SET). Использует SCAN
// (не KEYS) чтобы не блокировать Redis на большом keyspace.
func (s *CachedEventStore) invalidateUser(ctx context.Context, userID string) {
	if s.Cache == nil || s.Cache.Client == nil {
		return
	}
	pattern := s.Cache.Prefix + ":*:" + userID + ":*"
	iter := s.Cache.Client.Scan(ctx, 0, pattern, 100).Iterator()
	keys := []string{}
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if iter.Err() == nil && len(keys) > 0 {
		_ = s.Cache.Client.Del(ctx, keys...).Err()
	}
}

func (s *CachedEventStore) AggregateByCategory(ctx context.Context, userID string, since time.Time) (map[string]uint64, error) {
	key := fmt.Sprintf("agg:%s:%d", userID, since.Unix())
	var out map[string]uint64
	if hit, _ := s.Cache.GetJSON(ctx, key, &out); hit {
		return out, nil
	}
	out, err := s.Inner.AggregateByCategory(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if err := s.Cache.SetJSON(ctx, key, out, ttlAgg); err != nil {
		s.Logger.Debug("cache set failed", zap.String("key", key), zap.Error(err))
	}
	return out, nil
}

func (s *CachedEventStore) AggregateByCategoryBulk(ctx context.Context, userIDs []string, since time.Time) (map[string]map[string]uint64, error) {
	// Стабильный key из sorted userIDs — порядок caller'а не важен.
	sorted := append([]string(nil), userIDs...)
	sort.Strings(sorted)
	key := fmt.Sprintf("aggbulk:%s:%d", strings.Join(sorted, ","), since.Unix())
	var out map[string]map[string]uint64
	if hit, _ := s.Cache.GetJSON(ctx, key, &out); hit {
		return out, nil
	}
	out, err := s.Inner.AggregateByCategoryBulk(ctx, userIDs, since)
	if err != nil {
		return nil, err
	}
	if err := s.Cache.SetJSON(ctx, key, out, ttlBulk); err != nil {
		s.Logger.Debug("cache set failed", zap.String("key", key), zap.Error(err))
	}
	return out, nil
}

func (s *CachedEventStore) LanguageBreakdown(ctx context.Context, userID string, since time.Time) ([]LangCell, error) {
	key := fmt.Sprintf("lang:%s:%d", userID, since.Unix())
	var out []LangCell
	if hit, _ := s.Cache.GetJSON(ctx, key, &out); hit {
		return out, nil
	}
	out, err := s.Inner.LanguageBreakdown(ctx, userID, since)
	if err != nil {
		return nil, err
	}
	if err := s.Cache.SetJSON(ctx, key, out, ttlLang); err != nil {
		s.Logger.Debug("cache set failed", zap.String("key", key), zap.Error(err))
	}
	return out, nil
}

func (s *CachedEventStore) DailyTrend(ctx context.Context, userID string, since time.Time, tz string) ([]TrendPoint, error) {
	key := fmt.Sprintf("trend:%s:%d:%s", userID, since.Unix(), tz)
	var out []TrendPoint
	if hit, _ := s.Cache.GetJSON(ctx, key, &out); hit {
		return out, nil
	}
	out, err := s.Inner.DailyTrend(ctx, userID, since, tz)
	if err != nil {
		return nil, err
	}
	if err := s.Cache.SetJSON(ctx, key, out, ttlTrend); err != nil {
		s.Logger.Debug("cache set failed", zap.String("key", key), zap.Error(err))
	}
	return out, nil
}

func (s *CachedEventStore) Heatmap(ctx context.Context, userID string, since time.Time, tz string) ([]HeatmapCell, error) {
	key := fmt.Sprintf("heatmap:%s:%d:%s", userID, since.Unix(), tz)
	var out []HeatmapCell
	if hit, _ := s.Cache.GetJSON(ctx, key, &out); hit {
		return out, nil
	}
	out, err := s.Inner.Heatmap(ctx, userID, since, tz)
	if err != nil {
		return nil, err
	}
	if err := s.Cache.SetJSON(ctx, key, out, ttlHeatmap); err != nil {
		s.Logger.Debug("cache set failed", zap.String("key", key), zap.Error(err))
	}
	return out, nil
}
