// cache.go — Redis-backed read-through cache для public GET /v1/content/:slug.
//
// Design:
//   * Один key на (slug, requested_locale). Если client запросил `kk` и
//     fallback'нулись на `en`, ключ всё равно `eop:content:<slug>:kk` —
//     значение содержит effective locale внутри Entry. Это даёт точечный
//     invalidation per-locale + bust всех locales через InvalidateSlug
//     (на admin write — самый безопасный invariant).
//   * Cache hit = serve immediately. Cache miss = fetch from Store, write
//     back, return.
//   * Если Redis nil/down — graceful: Lookup возвращает (nil, false, nil),
//     Store/Invalidate — no-op.
//
// TTL: 5 минут по умолчанию (matches public Cache-Control max-age=300).

package content

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// DefaultCacheTTL — public read TTL = 5 минут (matches Scenario 7 в
// cms-content-render.md: Redis TTL = 300s).
const DefaultCacheTTL = 5 * time.Minute

// cacheKeyPrefix — namespaced под `eop:content:` чтобы не пересекаться с
// другими redis namespaces (`eop:cache:*`, `eop:webauthn:*`).
const cacheKeyPrefix = "eop:content:"

// Entry — сериализуемый payload в Redis. Handler отдаёт его напрямую в
// response без re-roundtrip к DB.
type Entry struct {
	Slug          string          `json:"slug"`
	Locale        string          `json:"locale"`         // effective locale (после fallback resolution)
	Content       json.RawMessage `json:"content"`
	SchemaVersion int             `json:"schema_version"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Source        string          `json:"source"`         // direct | locale_fallback:<l>
	ETag          string          `json:"etag"`
}

// Cache — обвёртка над redis.Client. Nil client => no-op (graceful degradation).
type Cache struct {
	Client *redis.Client
}

// NewCache — конструктор. Nil client допустим (no-op cache).
func NewCache(client *redis.Client) *Cache {
	return &Cache{Client: client}
}

// keyFor — `eop:content:<slug>:<locale>`.
func (c *Cache) keyFor(slug, locale string) string {
	return cacheKeyPrefix + slug + ":" + locale
}

// slugScanPattern — для InvalidateSlug. Matches все locale-ключи slug'а.
func (c *Cache) slugScanPattern(slug string) string {
	return cacheKeyPrefix + slug + ":*"
}

// Lookup читает cache entry. Возвращает (entry, hit, err). hit=false +
// err=nil = clean miss (нормальный path). err != nil = transport error.
func (c *Cache) Lookup(ctx context.Context, slug, locale string) (*Entry, bool, error) {
	if c == nil || c.Client == nil {
		return nil, false, nil
	}
	raw, err := c.Client.Get(ctx, c.keyFor(slug, locale)).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	var e Entry
	if err := json.Unmarshal(raw, &e); err != nil {
		// Corrupt entry (e.g. left over from prior serialization scheme)
		// → treat as miss; caller refetches from Postgres + overwrites.
		return nil, false, nil
	}
	return &e, true, nil
}

// Store записывает entry с TTL. ttl<=0 => DefaultCacheTTL. Nil entry — no-op.
func (c *Cache) Store(ctx context.Context, slug, locale string, entry *Entry, ttl time.Duration) error {
	if c == nil || c.Client == nil {
		return nil
	}
	if entry == nil {
		return nil
	}
	if ttl <= 0 {
		ttl = DefaultCacheTTL
	}
	body, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return c.Client.Set(ctx, c.keyFor(slug, locale), body, ttl).Err()
}

// Invalidate — точечный DEL для одного (slug, locale).
func (c *Cache) Invalidate(ctx context.Context, slug, locale string) error {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client.Del(ctx, c.keyFor(slug, locale)).Err()
}

// InvalidateSlug — bust все locale-ключи slug'а. Используется на admin
// write path (publish/draft/delete): когда `en` row обновился, кеш kk/ru/es
// мог содержать fallback'нувшийся `en`-контент — blast everything.
//
// Стратегия: SCAN по pattern + variadic DEL. Для miniredis (тесты) и
// production Redis работает одинаково.
func (c *Cache) InvalidateSlug(ctx context.Context, slug string) error {
	if c == nil || c.Client == nil {
		return nil
	}
	pattern := c.slugScanPattern(slug)
	var cursor uint64
	for {
		keys, next, err := c.Client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.Client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		if next == 0 {
			break
		}
		cursor = next
	}
	return nil
}
