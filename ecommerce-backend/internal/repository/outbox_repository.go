package repository

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

type OutboxRepository interface {
	CreateTx(ctx context.Context, tx *gorm.DB, event *model.OutboxEvent) error
	FetchPending(ctx context.Context, limit int) ([]model.OutboxEvent, error)
	MarkProcessed(ctx context.Context, id uint) error
	MarkFailed(ctx context.Context, id uint, errMessage string) error
}

type outboxRepository struct {
	db *gorm.DB
}

func NewOutboxRepository(db *gorm.DB) OutboxRepository {
	return &outboxRepository{db: db}
}

func (r *outboxRepository) CreateTx(ctx context.Context, tx *gorm.DB, event *model.OutboxEvent) error {
	db := tx
	if db == nil {
		db = r.db
	}
	if err := db.WithContext(ctx).Create(event).Error; err != nil {
		return fmt.Errorf("repository: create outbox event: %w", err)
	}
	return nil
}

func (r *outboxRepository) FetchPending(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	var events []model.OutboxEvent
	err := r.db.WithContext(ctx).
		Where("status = ?", model.OutboxStatusPending).
		Order("created_at ASC").
		Limit(limit).
		Find(&events).Error

	if err != nil {
		return nil, fmt.Errorf("repository: fetch pending outbox events: %w", err)
	}
	return events, nil
}

func (r *outboxRepository) MarkProcessed(ctx context.Context, id uint) error {
	now := time.Now()
	err := r.db.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":       model.OutboxStatusProcessed,
			"processed_at": &now,
		}).Error

	if err != nil {
		return fmt.Errorf("repository: mark outbox event processed: %w", err)
	}
	return nil
}

func (r *outboxRepository) MarkFailed(ctx context.Context, id uint, errMessage string) error {
	err := r.db.WithContext(ctx).
		Model(&model.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.OutboxStatusFailed,
			"retry_count": gorm.Expr("retry_count + 1"),
			"last_error":  errMessage,
		}).Error

	if err != nil {
		return fmt.Errorf("repository: mark outbox event failed: %w", err)
	}
	return nil
}
