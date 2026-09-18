// Package cache sets up the Redis client shared by caching (product/category
// cache-aside) and, optionally, the rate limiter.
package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// Options configures the Redis client connection.
type Options struct {
	Addr     string
	Password string
	DB       int
}

// Connect opens a Redis client and verifies connectivity with PING, failing
// fast if Redis is unreachable at startup.
func Connect(ctx context.Context, opts Options) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Password: opts.Password,
		DB:       opts.DB,
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("cache: connect: %w", err)
	}

	return client, nil
}
