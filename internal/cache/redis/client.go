package redis

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/subipraNuvem/url-shortener/internal/cache"
	"github.com/subipraNuvem/url-shortener/internal/config"
)

type Client struct {
	rdb        *goredis.Client
	defaultTTL time.Duration
	once       sync.Once
}

var (
	instance     *Client
	instanceOnce sync.Once
)

func NewClient(cfg *config.Config) *Client {
	instanceOnce.Do(func() {
		instance = &Client{
			defaultTTL: time.Duration(cfg.CacheDefaultTTLMillis) * time.Millisecond,
		}
	})
	return instance
}

func (c *Client) Connect(ctx context.Context, addr string) error {
	var connErr error
	c.once.Do(func() {
		c.rdb = goredis.NewClient(&goredis.Options{Addr: addr})
		err := c.rdb.Ping(ctx).Err()
		if err != nil {
			connErr = fmt.Errorf("redis connect: %w", err)
			return
		}
		slog.InfoContext(ctx, "redis connected", "addr", addr)
	})
	return connErr
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", nil
	}
	return val, err
}

func (c *Client) Set(ctx context.Context, key string, value string, opts ...cache.SetOption) error {
	o := &cache.SetOptions{TTL: c.defaultTTL}
	for _, opt := range opts {
		opt(o)
	}
	return c.rdb.Set(ctx, key, value, o.TTL).Err()
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return c.rdb.Del(ctx, key).Err()
}

func (c *Client) IncrementCounter(ctx context.Context, key string) (int64, error) {
	return c.rdb.Incr(ctx, key).Result()
}

func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

func (c *Client) Close() error {
	return c.rdb.Close()
}
