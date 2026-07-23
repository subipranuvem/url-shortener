package cache

import (
	"context"
	"time"
)

type SetOption func(*SetOptions)

type SetOptions struct {
	TTL time.Duration
}

func WithTTL(ttl time.Duration) SetOption {
	return func(o *SetOptions) { o.TTL = ttl }
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, opts ...SetOption) error
	Delete(ctx context.Context, key string) error
	IncrementCounter(ctx context.Context, key string) (int64, error)
	Ping(ctx context.Context) error
}
