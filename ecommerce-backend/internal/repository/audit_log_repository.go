package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// AuditLogRepository persists audit trail entries for sensitive actions
// (see PLANNING.md §2A). It exposes only Create — audit logs are
// write-once and never queried back through the API yet, so no read
// methods exist until a real consumer (admin panel) needs one.
type AuditLogRepository interface {
	Create(ctx context.Context, entry *model.AuditLog) error
}

type auditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository builds an AuditLogRepository backed by db.
func NewAuditLogRepository(db *gorm.DB) AuditLogRepository {
	return &auditLogRepository{db: db}
}

func (r *auditLogRepository) Create(ctx context.Context, entry *model.AuditLog) error {
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return fmt.Errorf("repository: create audit log: %w", err)
	}
	return nil
}
