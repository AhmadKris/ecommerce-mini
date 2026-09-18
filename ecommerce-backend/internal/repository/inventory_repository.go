package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ecommerce-backend/internal/model"
)

// ErrProductNotFound is returned by Adjust when the target product doesn't
// exist.
var ErrProductNotFound = errors.New("product not found")

// Adjust reuses ErrInsufficientStock from order_repository.go (same
// package) when an "out" adjustment would take stock below zero — same
// concept, and checkout/inventory adjustment are the only two places stock
// actually changes, so one sentinel for both is correct, not a shortcut.

// InventoryRepository records stock changes against a product, keeping
// Product.Stock and the inventory_movements audit trail consistent.
type InventoryRepository interface {
	Adjust(ctx context.Context, actorID uint, req model.AdjustInventoryRequest) (*model.InventoryMovement, error)
	ListByProductID(ctx context.Context, productID uint, page, limit int) ([]model.InventoryMovement, int64, error)
}

type inventoryRepository struct {
	db *gorm.DB
}

// NewInventoryRepository builds an InventoryRepository backed by db.
func NewInventoryRepository(db *gorm.DB) InventoryRepository {
	return &inventoryRepository{db: db}
}

// Adjust locks the product row, computes its new stock from req, and
// writes both the updated stock and the movement record in one
// transaction — the same row-lock-then-write pattern as
// OrderRepository.Checkout, for the same reason: two concurrent
// adjustments on the same product must serialize, not race.
func (r *inventoryRepository) Adjust(ctx context.Context, actorID uint, req model.AdjustInventoryRequest) (*model.InventoryMovement, error) {
	var movement model.InventoryMovement

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var product model.Product
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, req.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProductNotFound
			}
			return err
		}

		before := product.Stock
		after := before
		switch req.Type {
		case model.InventoryMovementTypeIn:
			after = before + req.Quantity
		case model.InventoryMovementTypeOut:
			after = before - req.Quantity
		case model.InventoryMovementTypeCorrection:
			after = req.Quantity
		}
		if after < 0 {
			return fmt.Errorf("%w: %s", ErrInsufficientStock, product.Name)
		}

		product.Stock = after
		if err := tx.Save(&product).Error; err != nil {
			return err
		}

		movement = model.InventoryMovement{
			ProductID:      req.ProductID,
			Type:           req.Type,
			Quantity:       req.Quantity,
			BeforeQuantity: before,
			AfterQuantity:  after,
			Reason:         req.Reason,
			Reference:      req.Reference,
			CreatedBy:      actorID,
		}
		return tx.Create(&movement).Error
	})
	if err != nil {
		return nil, fmt.Errorf("repository: adjust inventory: %w", err)
	}

	return &movement, nil
}

func (r *inventoryRepository) ListByProductID(ctx context.Context, productID uint, page, limit int) ([]model.InventoryMovement, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.InventoryMovement{}).
		Where("product_id = ?", productID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repository: count inventory movements: %w", err)
	}

	var movements []model.InventoryMovement
	err := r.db.WithContext(ctx).
		Where("product_id = ?", productID).
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&movements).Error
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list inventory movements: %w", err)
	}

	return movements, total, nil
}
