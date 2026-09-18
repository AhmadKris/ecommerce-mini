package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ecommerce-backend/internal/model"
)

// ErrEmptyCart means checkout was attempted with no items in the cart.
var ErrEmptyCart = errors.New("cart is empty")

// ErrInsufficientStock means a product's stock could not cover the
// quantity in the cart at checkout time. Separate from service.ErrInsufficientStock
// (same concept, different layer) — service detects this one via errors.Is
// and translates it to *apperror.AppError, same pattern as ErrEmailTaken.
var ErrInsufficientStock = errors.New("insufficient stock")

// flatShippingCost is a placeholder flat-rate shipping fee applied to every
// order — there is no courier/rate-table integration yet (Fase 2/3, see
// PLANNING.md). It is computed here, not in the frontend, so total_amount
// stays the single source of truth for what an order actually costs.
const flatShippingCost = 25000

// OrderRepository turns a user's cart into an order.
type OrderRepository interface {
	Checkout(ctx context.Context, userID uint, shippingAddress string) (*model.Order, error)
	ListByUserID(ctx context.Context, userID uint, page, limit int) ([]model.Order, int64, error)
}

type orderRepository struct {
	db *gorm.DB
}

// NewOrderRepository builds an OrderRepository backed by db.
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

// Checkout atomically turns userID's cart into an order: for each cart
// item it locks the product row (SELECT ... FOR UPDATE), checks stock,
// decrements it, then creates the order + order items + a placeholder
// payment, and clears the cart — all inside one transaction, so a crash or
// error midway leaves nothing partially applied.
//
// The FOR UPDATE lock is what actually prevents the race CLAUDE.md calls
// out ("cek stok di dalam transaction yang sama dengan pengurangan stok"):
// without it, two concurrent checkouts for the last unit of a product could
// both read stock=1, both pass the check, and both decrement — overselling
// by one. With it, the second transaction's lock acquisition blocks until
// the first commits (releasing the lock with the already-decremented
// value), so the second transaction's stock read is guaranteed current.
func (r *orderRepository) Checkout(ctx context.Context, userID uint, shippingAddress string) (*model.Order, error) {
	var order model.Order

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cart model.Cart
		if err := tx.Where("user_id = ?", userID).First(&cart).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrEmptyCart
			}
			return err
		}

		var cartItems []model.CartItem
		if err := tx.Where("cart_id = ?", cart.ID).Find(&cartItems).Error; err != nil {
			return err
		}
		if len(cartItems) == 0 {
			return ErrEmptyCart
		}

		order = model.Order{
			UserID:          userID,
			Status:          model.OrderStatusPending,
			ShippingCost:    flatShippingCost,
			ShippingAddress: shippingAddress,
		}
		orderItems := make([]model.OrderItem, 0, len(cartItems))

		for _, cartItem := range cartItems {
			var product model.Product
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, cartItem.ProductID).Error; err != nil {
				return err
			}
			if product.Stock < cartItem.Quantity {
				return fmt.Errorf("%w: %s", ErrInsufficientStock, product.Name)
			}

			product.Stock -= cartItem.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return err
			}

			order.TotalAmount += product.Price * float64(cartItem.Quantity)
			orderItems = append(orderItems, model.OrderItem{
				ProductID:       product.ID,
				Quantity:        cartItem.Quantity,
				PriceAtPurchase: product.Price,
				Product:         &product,
			})
		}

		order.TotalAmount += order.ShippingCost

		if err := tx.Create(&order).Error; err != nil {
			return err
		}
		for i := range orderItems {
			orderItems[i].OrderID = order.ID
		}
		if err := tx.Create(&orderItems).Error; err != nil {
			return err
		}
		order.Items = orderItems

		payment := model.Payment{OrderID: order.ID, Provider: "manual", Status: "pending"}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}

		if err := tx.Where("cart_id = ?", cart.ID).Delete(&model.CartItem{}).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("repository: checkout: %w", err)
	}

	return &order, nil
}

func (r *orderRepository) ListByUserID(ctx context.Context, userID uint, page, limit int) ([]model.Order, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repository: count orders: %w", err)
	}

	var orders []model.Order
	err := r.db.WithContext(ctx).
		Preload("Items.Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&orders).Error
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list orders: %w", err)
	}

	return orders, total, nil
}
