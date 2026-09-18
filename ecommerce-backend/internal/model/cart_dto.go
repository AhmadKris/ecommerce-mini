package model

// AddCartItemRequest is the payload for POST /api/cart/items.
type AddCartItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

// UpdateCartItemRequest is the payload for PUT /api/cart/items/:id. Unlike
// UpdateProductRequest, Quantity isn't a pointer — there's no "leave
// unchanged" case for a single-field update, and 0 is rejected (use DELETE
// to remove an item instead of PUT quantity=0).
type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

// CartItemResponse is one line of GET /api/cart's response, with Subtotal
// computed server-side so the client never has to trust/recompute pricing.
type CartItemResponse struct {
	ID       uint    `json:"id"`
	Product  Product `json:"product"`
	Quantity int     `json:"quantity"`
	Subtotal float64 `json:"subtotal"`
}

// CartResponse is the shape returned by GET /api/cart and every cart
// mutation endpoint, so the client always has the full up-to-date cart
// after an add/update/remove without a separate refetch.
type CartResponse struct {
	Items []CartItemResponse `json:"items"`
	Total float64            `json:"total"`
}
