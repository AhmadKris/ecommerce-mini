package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "ecommerce-backend/http"

// Tracing returns a Gin middleware that creates OpenTelemetry spans for incoming requests
// and sets X-Trace-ID response headers for distributed tracing.
func Tracing() gin.HandlerFunc {
	tracer := otel.Tracer(tracerName)

	return func(c *gin.Context) {
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		spanName := fmt.Sprintf("%s %s", c.Request.Method, path)
		ctx, span := tracer.Start(c.Request.Context(), spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("http.method", c.Request.Method),
				attribute.String("http.path", c.Request.URL.Path),
				attribute.String("http.user_agent", c.Request.UserAgent()),
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)

		if span.SpanContext().HasTraceID() {
			traceID := span.SpanContext().TraceID().String()
			c.Header("X-Trace-ID", traceID)
		}

		c.Next()

		status := c.Writer.Status()
		span.SetAttributes(
			attribute.Int("http.status_code", status),
		)
	}
}
