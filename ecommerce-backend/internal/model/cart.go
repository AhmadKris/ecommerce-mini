package model

import "time"

// Cart is one user's shopping cart. One-to-one with User (unique user_id).
type Cart struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

// CartItem is a product line in a cart. Unique on (cart_id, product_id) —
// adding a product already in the cart increments Quantity instead of
// creating a second row.
type CartItem struct {
	ID        uint `gorm:"primaryKey" json:"id"`
	CartID    uint `json:"cart_id"`
	ProductID uint `json:"product_id"`
	Quantity  int  `json:"quantity"`

	Product *Product `json:"product,omitempty"`
}
