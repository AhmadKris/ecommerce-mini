// Package logger configures the application's structured logger (log/slog)
// and propagates a per-request logger instance through context.Context.
package logger

import (
	"context"
	"log/slog"
	"os"
)

type contextKey struct{}

var loggerKey = contextKey{}

// New builds the base slog.Logger for the process: JSON output in
// staging/production for machine parsing, human-readable text in
// development for local debugging.
func New(env string) *slog.Logger {
	level := slog.LevelInfo
	opts := &slog.HandlerOptions{Level: level}

	var handler slog.Handler
	if env == "production" || env == "staging" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	return slog.New(handler)
}

// WithContext returns a new context carrying the given logger, typically
// enriched with a request ID by middleware before being attached.
func WithContext(ctx context.Context, log *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// FromContext retrieves the logger bound to ctx, falling back to
// slog.Default() if none was attached (e.g. background jobs outside the
// request lifecycle).
func FromContext(ctx context.Context) *slog.Logger {
	if log, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
		return log
	}
	return slog.Default()
}
