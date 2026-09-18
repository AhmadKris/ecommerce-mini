package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"ecommerce-backend/internal/logger"
)

const idempotencyHeader = "Idempotency-Key"

type idempotentRecord struct {
	Status int    `json:"status"`
	Body   []byte `json:"body"`
}

// bodyCapture wraps gin's ResponseWriter to record the status/body actually
// written, so Idempotency can cache it — the interface's other methods
// (Status, Size, Written, ...) forward to the embedded ResponseWriter
// unchanged.
type bodyCapture struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *bodyCapture) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *bodyCapture) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// Idempotency replays the first response for a request that repeats the
// same Idempotency-Key header (scoped per authenticated user) within ttl,
// instead of re-executing the handler. It exists for the checkout endpoint
// (see PLANNING.md §2C) — a client retrying after a dropped connection
// can't accidentally create two orders for the same cart. The header is
// opt-in: a request without it is never deduplicated.
func Idempotency(rdb *redis.Client, ttl time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader(idempotencyHeader)
		if key == "" {
			c.Next()
			return
		}

		ctx := c.Request.Context()
		userID, _ := UserIDFromContext(c)
		redisKey := fmt.Sprintf("idempotency:%d:%s:%s", userID, c.Request.Method, key)

		if cached, err := rdb.Get(ctx, redisKey).Result(); err == nil {
			var record idempotentRecord
			if jsonErr := json.Unmarshal([]byte(cached), &record); jsonErr == nil {
				c.Header("Idempotent-Replay", "true")
				c.Data(record.Status, "application/json", record.Body)
				c.Abort()
				return
			}
		}

		capture := &bodyCapture{ResponseWriter: c.Writer, body: &bytes.Buffer{}, status: 200}
		c.Writer = capture
		c.Next()

		if capture.status < 200 || capture.status >= 300 {
			return
		}
		data, err := json.Marshal(idempotentRecord{Status: capture.status, Body: capture.body.Bytes()})
		if err != nil {
			return
		}
		if err := rdb.Set(ctx, redisKey, data, ttl).Err(); err != nil {
			logger.FromContext(ctx).Error("idempotency: cache response failed", slog.Any("err", err))
		}
	}
}
