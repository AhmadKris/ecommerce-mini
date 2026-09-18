package model

// CheckoutRequest is the payload for POST /api/orders.
type CheckoutRequest struct {
	ShippingAddress string `json:"shipping_address" binding:"required,min=10,max=500"`
}

// OrderListQuery binds GET /api/orders query params.
type OrderListQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}
