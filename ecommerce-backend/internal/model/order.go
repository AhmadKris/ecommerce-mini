package model

import "time"

// Order status values, in their normal forward progression. Cancelled is
// reachable from Pending/Paid/Processing but not from Shipped/Delivered —
// see service.orderStatusTransitions for the full transition table.
const (
	OrderStatusPending    = "pending"
	OrderStatusPaid       = "paid"
	OrderStatusProcessing = "processing"
	OrderStatusShipped    = "shipped"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)

// Order is a completed checkout. TotalAmount and each OrderItem's
// PriceAtPurchase/ProductName are snapshotted at checkout time — they never
// change even if the product's live price/name does later.
type Order struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	UserID          uint      `json:"user_id"`
	Status          string    `json:"status"`
	TotalAmount     float64   `json:"total_amount"`
	ShippingCost    float64   `json:"shipping_cost"`
	ShippingAddress string    `json:"shipping_address"`
	CreatedAt       time.Time `json:"created_at"`

	Items []OrderItem `json:"items,omitempty"`
}

// OrderItem is one product line of an order, with price and name
// snapshotted at purchase time. ProductName is what the client should
// display — never Product.Name — so a later rename or deletion of the
// product can't silently rewrite historical order data (see
// .claude/CLAUDE.md Known Issues for the bug this fixes). Product is still
// preloaded for fields that are fine to show live (e.g. current image_url),
// but Product itself may be nil if the product was hard-deleted.
type OrderItem struct {
	ID              uint    `gorm:"primaryKey" json:"id"`
	OrderID         uint    `json:"order_id"`
	ProductID       uint    `json:"product_id"`
	ProductName     string  `json:"product_name"`
	Quantity        int     `json:"quantity"`
	PriceAtPurchase float64 `json:"price_at_purchase"`

	Product *Product `json:"product,omitempty"`
}

// Payment is a placeholder payment record created at checkout — there is no
// real payment gateway integration yet (Fase 2/3, see PLANNING.md), so every
// payment is created with Provider "manual" and Status "pending".
type Payment struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	OrderID       uint       `json:"order_id"`
	Provider      string     `json:"provider"`
	Status        string     `json:"status"`
	TransactionID string     `json:"transaction_id"`
	PaidAt        *time.Time `json:"paid_at"`
}
