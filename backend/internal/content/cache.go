package content

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

const DefaultCacheTTL = 5 * time.Minute

const cacheKeyPrefix = "eop:content:"

type Entry struct {
	Slug          string          `json:"slug"`
	Locale        string          `json:"locale"`
	Content       json.RawMessage `json:"content"`
	SchemaVersion int             `json:"schema_version"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	UpdatedAt     time.Time       `json:"updated_at"`
	Source        string          `json:"source"`
	ETag          string          `json:"etag"`
}

type Cache struct {
	Client *redis.Client
}

func NewCache(client *redis.Client) *Cache {
	return &Cache{Client: client}
}

func (c *Cache) keyFor(slug, locale string) string {
	return cacheKeyPrefix + slug + ":" + locale
}

func (c *Cache) slugScanPattern(slug string) string {
	return cacheKeyPrefix + slug + ":*"
}

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

		return nil, false, nil
	}
	return &e, true, nil
}

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

func (c *Cache) Invalidate(ctx context.Context, slug, locale string) error {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client.Del(ctx, c.keyFor(slug, locale)).Err()
}

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
