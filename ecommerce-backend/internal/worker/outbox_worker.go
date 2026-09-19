package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

type OutboxWorker struct {
	outboxRepo repository.OutboxRepository
	pollInterval time.Duration
}

func NewOutboxWorker(outboxRepo repository.OutboxRepository, pollInterval time.Duration) *OutboxWorker {
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	return &OutboxWorker{
		outboxRepo:   outboxRepo,
		pollInterval: pollInterval,
	}
}

// Start runs the worker event processing loop until ctx is cancelled (graceful shutdown).
func (w *OutboxWorker) Start(ctx context.Context) {
	slog.Info("Outbox background worker started", "poll_interval", w.pollInterval.String())
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Outbox background worker shutting down gracefully")
			return
		case <-ticker.C:
			w.processPending(ctx)
		}
	}
}

func (w *OutboxWorker) processPending(ctx context.Context) {
	events, err := w.outboxRepo.FetchPending(ctx, 10)
	if err != nil {
		slog.Error("Outbox worker failed to fetch pending events", "error", err)
		return
	}

	for _, event := range events {
		if err := w.handleEvent(ctx, &event); err != nil {
			slog.Error("Outbox worker failed to handle event", "event_id", event.ID, "event_type", event.EventType, "error", err)
			_ = w.outboxRepo.MarkFailed(ctx, event.ID, err.Error())
		} else {
			slog.Info("Outbox worker successfully processed event", "event_id", event.ID, "event_type", event.EventType)
			_ = w.outboxRepo.MarkProcessed(ctx, event.ID)
		}
	}
}

func (w *OutboxWorker) handleEvent(ctx context.Context, event *model.OutboxEvent) error {
	switch event.EventType {
	case "order.created":
		var payload map[string]interface{}
		if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
			return err
		}
		// Async task simulation: send order confirmation email & trigger fulfillment queue
		slog.Info("Simulating order creation async notification", "order_id", payload["order_id"], "user_id", payload["user_id"])
		return nil
	default:
		slog.Info("Processing generic outbox event", "event_type", event.EventType)
		return nil
	}
}
