package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/logger"
)

// RequestLogging logs one structured line per request. 4xx responses are
// normal traffic and logged at Info; only 5xx responses are logged at Error.
func RequestLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		status := c.Writer.Status()
		log := logger.FromContext(c.Request.Context())
		attrs := []any{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", status),
			slog.Duration("latency", time.Since(start)),
		}

		if status >= 500 {
			log.Error("request completed", attrs...)
		} else {
			log.Info("request completed", attrs...)
		}
	}
}
