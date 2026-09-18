package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RefreshBlacklist tracks refresh token jti values that must be rejected —
// either because they were already rotated (single-use enforcement) or
// explicitly revoked. Backed by Redis so entries expire on their own once
// the token would have expired anyway.
type RefreshBlacklist struct {
	redis *redis.Client
}

// NewRefreshBlacklist builds a RefreshBlacklist backed by client.
func NewRefreshBlacklist(client *redis.Client) *RefreshBlacklist {
	return &RefreshBlacklist{redis: client}
}

// Add blacklists jti for ttl. A non-positive ttl means the token is already
// expired and needs no entry — expiry alone rejects it.
func (b *RefreshBlacklist) Add(ctx context.Context, jti string, ttl time.Duration) error {
	if ttl <= 0 {
		return nil
	}
	if err := b.redis.Set(ctx, blacklistKey(jti), "1", ttl).Err(); err != nil {
		return fmt.Errorf("auth: blacklist refresh token: %w", err)
	}
	return nil
}

// IsBlacklisted reports whether jti has been blacklisted.
func (b *RefreshBlacklist) IsBlacklisted(ctx context.Context, jti string) (bool, error) {
	count, err := b.redis.Exists(ctx, blacklistKey(jti)).Result()
	if err != nil {
		return false, fmt.Errorf("auth: check refresh token blacklist: %w", err)
	}
	return count > 0, nil
}

func blacklistKey(jti string) string {
	return "refresh_blacklist:" + jti
}
