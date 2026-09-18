package model

import "time"

// OrderStatusPending is the only status an order can have in Fase 1 — no
// status-transition endpoint exists yet (see Known Issues in
// .claude/CLAUDE.md), so orders never leave this state today.
const OrderStatusPending = "pending"

// Order is a completed checkout. TotalAmount and each OrderItem's
// PriceAtPurchase are snapshotted at checkout time — they never change even
// if the product's live price does later.
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

// OrderItem is one product line of an order, with the price snapshotted at
// purchase time.
type OrderItem struct {
	ID              uint    `gorm:"primaryKey" json:"id"`
	OrderID         uint    `json:"order_id"`
	ProductID       uint    `json:"product_id"`
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
