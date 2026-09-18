package service

import (
	"context"
	"errors"
	"fmt"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// InventoryService implements stock adjustments and their movement
// history. No separate audit_logs entry is written here — the
// inventory_movements row created by Adjust already is the dedicated audit
// trail for stock changes (before/after quantity, reason, actor), a
// different purpose from audit_logs' generic per-resource log.
type InventoryService struct {
	inventoryRepo repository.InventoryRepository
}

// NewInventoryService builds an InventoryService with its dependencies.
func NewInventoryService(inventoryRepo repository.InventoryRepository) *InventoryService {
	return &InventoryService{inventoryRepo: inventoryRepo}
}

// Adjust applies a stock change (see model.AdjustInventoryRequest) and
// records it. All locking/atomicity lives in the repository.
func (s *InventoryService) Adjust(ctx context.Context, actorID uint, req model.AdjustInventoryRequest) (*model.InventoryMovement, error) {
	movement, err := s.inventoryRepo.Adjust(ctx, actorID, req)
	if err != nil {
		if errors.Is(err, repository.ErrProductNotFound) {
			return nil, apperror.NotFound("Produk tidak ditemukan", nil)
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			return nil, apperror.Conflict("Stok tidak mencukupi untuk penyesuaian ini", err)
		}
		return nil, apperror.Internal(fmt.Errorf("service: adjust inventory: %w", err))
	}
	return movement, nil
}

// ListMovements returns a page of a product's stock movement history,
// newest first.
func (s *InventoryService) ListMovements(ctx context.Context, productID uint, query model.InventoryMovementListQuery) ([]model.InventoryMovement, model.Meta, error) {
	page := query.Page
	if page < 1 {
		page = defaultPage
	}
	limit := query.Limit
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	movements, total, err := s.inventoryRepo.ListByProductID(ctx, productID, page, limit)
	if err != nil {
		return nil, model.Meta{}, apperror.Internal(fmt.Errorf("service: list inventory movements: %w", err))
	}

	return movements, model.NewMeta(page, limit, total), nil
}
