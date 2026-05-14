package cache

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Client *redis.Client
	Prefix string
}

const defaultPrefix = "eop:cache"

func New(ctx context.Context, addr string) (*Cache, error) {
	if addr == "" {
		return nil, errors.New("redis addr empty")
	}
	opts, err := buildOptions(addr)
	if err != nil {
		return nil, err
	}
	cli := redis.NewClient(opts)
	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := cli.Ping(pingCtx).Err(); err != nil {
		_ = cli.Close()
		return nil, err
	}
	return &Cache{Client: cli, Prefix: defaultPrefix}, nil
}

func buildOptions(addr string) (*redis.Options, error) {
	const (
		poolSize     = 10
		readTimeout  = 500 * time.Millisecond
		writeTimeout = 500 * time.Millisecond
		dialTimeout  = 2 * time.Second
	)
	if strings.Contains(addr, "://") {
		opts, err := redis.ParseURL(addr)
		if err != nil {
			return nil, err
		}
		opts.PoolSize = poolSize
		opts.ReadTimeout = readTimeout
		opts.WriteTimeout = writeTimeout
		opts.DialTimeout = dialTimeout
		return opts, nil
	}
	return &redis.Options{
		Addr:         addr,
		PoolSize:     poolSize,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		DialTimeout:  dialTimeout,
	}, nil
}

func (c *Cache) key(suffix string) string {
	if c == nil || c.Client == nil {
		return ""
	}
	return c.Prefix + ":" + suffix
}

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

func (c *Cache) Close() error {
	if c == nil || c.Client == nil {
		return nil
	}
	return c.Client.Close()
}
