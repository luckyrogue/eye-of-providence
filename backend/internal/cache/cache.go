// Package cache — thin Redis-backed KV cache для read-heavy aggregations.
//
// Design: decorator pattern на EventStore. Если Redis доступен (cfg.RedisAddr
// reachable), оборачиваем CH-store в CachedEventStore. Cache miss → fetch +
// store; cache hit → return сразу.
//
// Cache keys: "{prefix}:{userID}:{params...}" — например
//   - eop:cache:agg:abc-123:since=1234567890
//   - eop:cache:lang:abc-123:days=30
//
// TTL — короткие (5-10 min) т.к. dashboard not real-time:
//   - Aggregates / language breakdown → 10 min
//   - DailyTrend / Heatmap → 5 min (active dashboards refresh)
//
// Invalidation: TTL-only (no eventual-consistency penalty за couple-min
// staleness). Future: invalidate-on-ingest через PubSub. По нынешним 3
// founding companies overhead не оправдан.
package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache — thin абстракция над redis.Client. Skipped когда Client == nil
// (no-op fallback при отсутствии Redis). Методы JSON-marshal'ят value через
// generic Get/Set.
type Cache struct {
	Client *redis.Client
	Prefix string // "eop:cache" по умолчанию
}

const defaultPrefix = "eop:cache"

// New — конструктор. Тестит соединение через PING; если fails — возвращает
// nil + err, caller должен decide fallback (typically — оставить event store
// uncached).
func New(ctx context.Context, addr string) (*Cache, error) {
	if addr == "" {
		return nil, errors.New("redis addr empty")
	}
	cli := redis.NewClient(&redis.Options{
		Addr:         addr,
		PoolSize:     10,
		ReadTimeout:  500 * time.Millisecond,
		WriteTimeout: 500 * time.Millisecond,
		DialTimeout:  2 * time.Second,
	})
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := cli.Ping(pingCtx).Err(); err != nil {
		_ = cli.Close()
		return nil, err
	}
	return &Cache{Client: cli, Prefix: defaultPrefix}, nil
}

// key — глобальный prefix + user-scoped key. Keeps keys grouped per pre-fix
// для удобной FLUSH operations.
func (c *Cache) key(suffix string) string {
	if c == nil || c.Client == nil {
		return ""
	}
	return c.Prefix + ":" + suffix
}

// GetJSON — попытка прочесть key. miss → возвращает (false, nil) — это
// нормальный path, caller fetch'ит fresh + сохраняет.
func (c *Cache) GetJSON(ctx context.Context, key string, dst any) (bool, error) {
	if c == nil || c.Client == nil {
		return false, nil
	}
	raw, err := c.Client.Get(ctx, c.key(key)).Bytes()
	if errors.Is(err, redis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal(raw, dst); err != nil {
		return false, err
	}
	return true, nil
}

// SetJSON — фоном (non-fatal): если Redis temporarily down, мы только
// логируем. cache write failure не должен ломать запрос.
func (c *Cache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if c == nil || c.Client == nil {
		return nil
	}
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.Client.Set(ctx, c.key(key), body, ttl).Err()
}

// Close — освобождает connection pool. Caller вызывает на shutdown.
func (c *Cache) Close() error {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client.Close()
}
