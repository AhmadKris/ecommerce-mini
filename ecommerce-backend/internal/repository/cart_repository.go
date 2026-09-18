package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// CartRepository persists and retrieves a user's cart and its items.
type CartRepository interface {
	FindOrCreateByUserID(ctx context.Context, userID uint) (*model.Cart, error)
	FindItemByID(ctx context.Context, itemID uint) (*model.CartItem, error)
	FindItemByProductID(ctx context.Context, cartID, productID uint) (*model.CartItem, error)
	CreateItem(ctx context.Context, item *model.CartItem) error
	UpdateItemQuantity(ctx context.Context, item *model.CartItem) error
	DeleteItem(ctx context.Context, itemID uint) error
	ListItems(ctx context.Context, cartID uint) ([]model.CartItem, error)
}

type cartRepository struct {
	db *gorm.DB
}

// NewCartRepository builds a CartRepository backed by db.
func NewCartRepository(db *gorm.DB) CartRepository {
	return &cartRepository{db: db}
}

// FindOrCreateByUserID fetches userID's cart, lazily creating one on first
// use. A unique-violation on create means a concurrent request created the
// same user's cart first — that row is fetched and returned instead of
// erroring, the same race-safe pattern as user/product creation.
func (r *cartRepository) FindOrCreateByUserID(ctx context.Context, userID uint) (*model.Cart, error) {
	var cart model.Cart
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart).Error
	if err == nil {
		return &cart, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("repository: find cart: %w", err)
	}

	cart = model.Cart{UserID: userID}
	if err := r.db.WithContext(ctx).Create(&cart).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
			if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&cart).Error; err != nil {
				return nil, fmt.Errorf("repository: find cart after race: %w", err)
			}
			return &cart, nil
		}
		return nil, fmt.Errorf("repository: create cart: %w", err)
	}
	return &cart, nil
}

func (r *cartRepository) FindItemByID(ctx context.Context, itemID uint) (*model.CartItem, error) {
	var item model.CartItem
	err := r.db.WithContext(ctx).Preload("Product").First(&item, itemID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find cart item: %w", err)
	}
	return &item, nil
}

func (r *cartRepository) FindItemByProductID(ctx context.Context, cartID, productID uint) (*model.CartItem, error) {
	var item model.CartItem
	err := r.db.WithContext(ctx).
		Where("cart_id = ? AND product_id = ?", cartID, productID).
		First(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find cart item by product: %w", err)
	}
	return &item, nil
}

func (r *cartRepository) CreateItem(ctx context.Context, item *model.CartItem) error {
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return fmt.Errorf("repository: create cart item: %w", err)
	}
	return nil
}

func (r *cartRepository) UpdateItemQuantity(ctx context.Context, item *model.CartItem) error {
	if err := r.db.WithContext(ctx).Save(item).Error; err != nil {
		return fmt.Errorf("repository: update cart item: %w", err)
	}
	return nil
}

func (r *cartRepository) DeleteItem(ctx context.Context, itemID uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.CartItem{}, itemID).Error; err != nil {
		return fmt.Errorf("repository: delete cart item: %w", err)
	}
	return nil
}

func (r *cartRepository) ListItems(ctx context.Context, cartID uint) ([]model.CartItem, error) {
	var items []model.CartItem
	err := r.db.WithContext(ctx).
		Preload("Product").
		Where("cart_id = ?", cartID).
		Order("id").
		Find(&items).Error
	if err != nil {
		return nil, fmt.Errorf("repository: list cart items: %w", err)
	}
	return items, nil
}
