package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"ecommerce-backend/internal/apperror"
)

// RateLimit rejects a client (identified by IP) with 429 once it exceeds
// limit requests within window, counted as a fixed-window counter in Redis —
// shared across API instances, unlike an in-memory counter, so the limit
// still holds once this runs behind a load balancer (see PLANNING.md's
// stateless-service principle).
func RateLimit(rdb *redis.Client, keyPrefix string, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		key := fmt.Sprintf("ratelimit:%s:%s", keyPrefix, c.ClientIP())

		count, err := rdb.Incr(ctx, key).Result()
		if err != nil {
			// Redis being unreachable should not take auth entirely
			// offline — fail open rather than 500 every login attempt.
			c.Next()
			return
		}
		if count == 1 {
			rdb.Expire(ctx, key, window)
		}
		if count > int64(limit) {
			_ = c.Error(apperror.RateLimited("Terlalu banyak percobaan, coba lagi beberapa saat lagi"))
			c.Abort()
			return
		}

		c.Next()
	}
}
