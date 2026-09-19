package model

// CreateReviewRequest is the payload for POST /api/products/:slug/reviews.
// OrderID isn't in the request — the service looks up the caller's own
// delivered order for this product itself, so a client can't claim a
// purchase (or someone else's order) that isn't theirs.
type CreateReviewRequest struct {
	Rating int    `json:"rating" binding:"required,gte=1,lte=5"`
	Title  string `json:"title" binding:"required,min=3,max=255"`
	Body   string `json:"body" binding:"required,min=10,max=2000"`
}

// ReviewListQuery binds GET /api/products/:slug/reviews and
// GET /api/admin/reviews query params.
type ReviewListQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

// UpdateReviewStatusRequest is the payload for
// PATCH /api/admin/reviews/:id/status.
type UpdateReviewStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}
