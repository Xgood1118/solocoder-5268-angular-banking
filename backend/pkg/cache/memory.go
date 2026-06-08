package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type MemoryCache struct {
	data map[string]*cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value    string
	expireAt *time.Time
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{
		data: make(map[string]*cacheItem),
	}
}

func (m *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, ok := m.data[key]
	if !ok {
		return "", errors.New("redis: nil")
	}

	if item.expireAt != nil && time.Now().After(*item.expireAt) {
		return "", errors.New("redis: nil")
	}

	return item.value, nil
}

func (m *MemoryCache) GetFloat64(ctx context.Context, key string) (float64, error) {
	val, err := m.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	var f float64
	fmt.Sscanf(val, "%f", &f)
	return f, nil
}

func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expireAt *time.Time
	if expiration > 0 {
		t := time.Now().Add(expiration)
		expireAt = &t
	}

	valStr := fmt.Sprintf("%v", value)
	m.data[key] = &cacheItem{
		value:    valStr,
		expireAt: expireAt,
	}

	return nil
}

func (m *MemoryCache) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if item, ok := m.data[key]; ok {
		if item.expireAt == nil || time.Now().Before(*item.expireAt) {
			return false, nil
		}
	}

	var expireAt *time.Time
	if expiration > 0 {
		t := time.Now().Add(expiration)
		expireAt = &t
	}

	valStr := fmt.Sprintf("%v", value)
	m.data[key] = &cacheItem{
		value:    valStr,
		expireAt: expireAt,
	}

	return true, nil
}

func (m *MemoryCache) IncrByFloat(ctx context.Context, key string, value float64) (float64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var current float64
	item, ok := m.data[key]
	if ok && (item.expireAt == nil || time.Now().Before(*item.expireAt)) {
		fmt.Sscanf(item.value, "%f", &current)
	}

	newVal := current + value
	m.data[key] = &cacheItem{
		value:    fmt.Sprintf("%f", newVal),
		expireAt: func() *time.Time {
			if item != nil && item.expireAt != nil {
				return item.expireAt
			}
			return nil
		}(),
	}

	return newVal, nil
}

func (m *MemoryCache) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.data[key]
	if !ok {
		return false, nil
	}

	if item.expireAt != nil && time.Now().After(*item.expireAt) {
		return false, nil
	}

	t := time.Now().Add(expiration)
	item.expireAt = &t
	return true, nil
}

func (m *MemoryCache) ExpireAt(ctx context.Context, key string, tm time.Time) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	item, ok := m.data[key]
	if !ok {
		return false, nil
	}

	item.expireAt = &tm
	return true, nil
}

func (m *MemoryCache) Del(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, key := range keys {
		delete(m.data, key)
	}
	return nil
}

func (m *MemoryCache) Ping(ctx context.Context) error {
	return nil
}

func (m *MemoryCache) Close() error {
	return nil
}
