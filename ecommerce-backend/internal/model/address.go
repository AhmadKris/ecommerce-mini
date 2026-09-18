package model

import "time"

// Address is a saved shipping address in a user's address book — separate
// from Order.ShippingAddress, which is a frozen text snapshot taken from an
// Address (or typed free-form) at checkout time. Editing an Address never
// changes past orders.
type Address struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `json:"user_id"`
	RecipientName string    `json:"recipient_name"`
	Phone         string    `json:"phone"`
	AddressLine   string    `json:"address_line"`
	City          string    `json:"city"`
	Province      string    `json:"province"`
	PostalCode    string    `json:"postal_code"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
