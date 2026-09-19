package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// PasswordResetStore issues and redeems single-use password reset tokens,
// backed by Redis so entries expire on their own — no cleanup job needed,
// same reasoning as RefreshBlacklist. Only the token's SHA-256 hash is ever
// stored; the raw token exists solely in the response handed to the caller
// of GenerateToken (and, in this project's current no-email-provider setup,
// in the server log it's delivered through — see AuthService.ForgotPassword).
// A Redis dump leaking the hash can't be used to reset anyone's password.
type PasswordResetStore struct {
	redis *redis.Client
}

// NewPasswordResetStore builds a PasswordResetStore backed by client.
func NewPasswordResetStore(client *redis.Client) *PasswordResetStore {
	return &PasswordResetStore{redis: client}
}

// GenerateToken creates a new opaque reset token for userID, valid for ttl,
// and returns the raw token to deliver to the user.
func (s *PasswordResetStore) GenerateToken(ctx context.Context, userID uint, ttl time.Duration) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("auth: generate password reset token: %w", err)
	}
	token := hex.EncodeToString(raw)

	if err := s.redis.Set(ctx, resetKey(token), userID, ttl).Err(); err != nil {
		return "", fmt.Errorf("auth: store password reset token: %w", err)
	}
	return token, nil
}

// Consume redeems token, returning the userID it was issued for and
// deleting it so it can't be used again. A zero userID with a nil error
// means the token was never issued, already used, or has expired.
func (s *PasswordResetStore) Consume(ctx context.Context, token string) (uint, error) {
	key := resetKey(token)

	userID, err := s.redis.Get(ctx, key).Uint64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("auth: read password reset token: %w", err)
	}

	if err := s.redis.Del(ctx, key).Err(); err != nil {
		return 0, fmt.Errorf("auth: delete password reset token: %w", err)
	}
	return uint(userID), nil
}

func resetKey(token string) string {
	hashed := sha256.Sum256([]byte(token))
	return "password_reset:" + hex.EncodeToString(hashed[:])
}
