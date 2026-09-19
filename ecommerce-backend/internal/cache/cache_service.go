package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService interface {
	Get(ctx context.Context, key string, dest interface{}) bool
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	InvalidatePattern(ctx context.Context, pattern string) error
}

type redisCacheService struct {
	client *redis.Client
}

func NewCacheService(client *redis.Client) CacheService {
	if client == nil {
		return &noopCacheService{}
	}
	return &redisCacheService{client: client}
}

func (c *redisCacheService) Get(ctx context.Context, key string, dest interface{}) bool {
	val, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err != redis.Nil {
			slog.Debug("Cache miss / get error", "key", key, "error", err)
		}
		return false
	}
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		slog.Error("Failed to unmarshal cached JSON", "key", key, "error", err)
		return false
	}
	return true
}

func (c *redisCacheService) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache: marshal value: %w", err)
	}
	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		slog.Error("Failed to set cache key", "key", key, "error", err)
		return err
	}
	return nil
}

func (c *redisCacheService) InvalidatePattern(ctx context.Context, pattern string) error {
	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("cache: scan keys for pattern %s: %w", pattern, err)
		}
		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				slog.Error("Failed to delete cache keys", "pattern", pattern, "error", err)
			}
		}
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	return nil
}

// noopCacheService fallback if Redis is disabled or unconfigured in dev/testing environment.
type noopCacheService struct{}

func (n *noopCacheService) Get(_ context.Context, _ string, _ interface{}) bool {
	return false
}

func (n *noopCacheService) Set(_ context.Context, _ string, _ interface{}, _ time.Duration) error {
	return nil
}

func (n *noopCacheService) InvalidatePattern(_ context.Context, _ string) error {
	return nil
}
