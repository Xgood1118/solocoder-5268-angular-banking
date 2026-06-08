package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	GetFloat64(ctx context.Context, key string) (float64, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	IncrByFloat(ctx context.Context, key string, value float64) (float64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) (bool, error)
	ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error)
	Del(ctx context.Context, keys ...string) error
	Ping(ctx context.Context) error
	Close() error
}
