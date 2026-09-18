package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"ecommerce-backend/internal/logger"
)

const requestIDHeader = "X-Request-ID"

// RequestID assigns a request ID (reusing the client's X-Request-ID header
// when present) and binds a request-scoped logger to the request context so
// downstream code can retrieve it via logger.FromContext.
func RequestID(base *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(requestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Writer.Header().Set(requestIDHeader, requestID)

		requestLogger := base.With(slog.String("request_id", requestID))
		ctx := logger.WithContext(c.Request.Context(), requestLogger)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
