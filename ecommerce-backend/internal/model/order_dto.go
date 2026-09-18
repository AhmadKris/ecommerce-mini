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

// UpdateOrderStatusRequest is the payload for
// PATCH /api/admin/orders/:id/status. The `oneof` tag rejects a typo'd or
// unknown status before it ever reaches the service's transition check.
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending paid processing shipped delivered cancelled"`
}
