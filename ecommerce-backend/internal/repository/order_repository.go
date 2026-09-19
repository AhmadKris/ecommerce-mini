package repository

import (
	"context"
	"encoding/json"
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
	Checkout(ctx context.Context, userID uint, shippingAddress string, promoCode string) (*model.Order, error)
	ListByUserID(ctx context.Context, userID uint, page, limit int) ([]model.Order, int64, error)
	ListAll(ctx context.Context, page, limit int) ([]model.Order, int64, error)
	FindByID(ctx context.Context, id uint) (*model.Order, error)
	UpdateStatus(ctx context.Context, id uint, status string) (*model.Order, error)
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
func (r *orderRepository) Checkout(ctx context.Context, userID uint, shippingAddress string, promoCode string) (*model.Order, error) {
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
			orderItem, subtotal, err := lockAndReserveStock(tx, cartItem)
			if err != nil {
				return err
			}
			order.TotalAmount += subtotal
			orderItems = append(orderItems, orderItem)
		}

		if err := applyPromoCode(tx, &order, promoCode); err != nil {
			return err
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

		// Written in the same transaction as the stock decrement, not as a
		// best-effort side effect afterwards — an audit trail that could
		// silently go missing on a partial failure defeats its own purpose.
		return writeCheckoutAuditLog(tx, order, userID, len(orderItems))
	})
	if err != nil {
		return nil, fmt.Errorf("repository: checkout: %w", err)
	}

	return &order, nil
}

// lockAndReserveStock locks cartItem's product row (SELECT ... FOR UPDATE),
// checks and decrements its stock, and returns the OrderItem it becomes —
// split out of Checkout purely to keep that function's cyclomatic
// complexity under the project's gocyclo limit; behavior is unchanged, tx
// is still the same transaction Checkout is running in.
func lockAndReserveStock(tx *gorm.DB, cartItem model.CartItem) (model.OrderItem, float64, error) {
	var product model.Product
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, cartItem.ProductID).Error; err != nil {
		return model.OrderItem{}, 0, err
	}
	if product.Stock < cartItem.Quantity {
		return model.OrderItem{}, 0, fmt.Errorf("%w: %s", ErrInsufficientStock, product.Name)
	}

	product.Stock -= cartItem.Quantity
	if err := tx.Save(&product).Error; err != nil {
		return model.OrderItem{}, 0, err
	}

	orderItem := model.OrderItem{
		ProductID:       product.ID,
		ProductName:     product.Name,
		Quantity:        cartItem.Quantity,
		PriceAtPurchase: product.Price,
		Product:         &product,
	}
	subtotal := product.Price * float64(cartItem.Quantity)
	return orderItem, subtotal, nil
}

// writeCheckoutAuditLog records the order.checkout audit entry — split out
// of Checkout for the same gocyclo reason as lockAndReserveStock/
// applyPromoCode, not because it's reused elsewhere.
func writeCheckoutAuditLog(tx *gorm.DB, order model.Order, userID uint, itemCount int) error {
	auditMetadata, err := json.Marshal(map[string]any{
		"total_amount":    order.TotalAmount,
		"shipping_cost":   order.ShippingCost,
		"discount_amount": order.DiscountAmount,
		"item_count":      itemCount,
	})
	if err != nil {
		return fmt.Errorf("marshal audit metadata: %w", err)
	}
	auditLog := model.AuditLog{
		ActorID:    userID,
		Action:     "order.checkout",
		Resource:   "order",
		ResourceID: order.ID,
		Metadata:   string(auditMetadata),
	}
	return tx.Create(&auditLog).Error
}

// applyPromoCode redeems promoCode against order's current TotalAmount
// (items-only subtotal at this point, before shipping is added) and, if
// valid, applies the discount to order in place. A no-op when promoCode is
// empty — split out of Checkout purely to keep its cyclomatic complexity
// under the project's gocyclo limit, same reasoning as lockAndReserveStock.
func applyPromoCode(tx *gorm.DB, order *model.Order, promoCode string) error {
	if promoCode == "" {
		return nil
	}
	promotion, discount, err := lockAndRedeemPromotion(tx, promoCode, order.TotalAmount)
	if err != nil {
		return err
	}
	order.DiscountAmount = discount
	order.PromotionID = &promotion.ID
	order.TotalAmount -= discount
	return nil
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

// FindByID returns a single order (with items preloaded), or nil if it
// doesn't exist.
func (r *orderRepository) FindByID(ctx context.Context, id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Preload("Items.Product").First(&order, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find order by id: %w", err)
	}
	return &order, nil
}

// UpdateStatus sets an order's status. Transition validity (e.g. a
// "delivered" order can't go back to "pending") is a business rule checked
// by the service layer, not here — this is a plain write.
func (r *orderRepository) UpdateStatus(ctx context.Context, id uint, status string) (*model.Order, error) {
	if err := r.db.WithContext(ctx).Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return nil, fmt.Errorf("repository: update order status: %w", err)
	}
	return r.FindByID(ctx, id)
}

// ListAll returns a page of every order across all users, newest first —
// backs the admin order list (order:read_all), unlike ListByUserID which is
// scoped to one customer's own history.
func (r *orderRepository) ListAll(ctx context.Context, page, limit int) ([]model.Order, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Order{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repository: count all orders: %w", err)
	}

	var orders []model.Order
	err := r.db.WithContext(ctx).
		Preload("Items.Product").
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&orders).Error
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list all orders: %w", err)
	}

	return orders, total, nil
}
