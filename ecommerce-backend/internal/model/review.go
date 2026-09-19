package model

import "time"

// Review moderation states. A new review always starts Pending — it's
// only visible on the public product page once an admin sets it Approved.
const (
	ReviewStatusPending  = "pending"
	ReviewStatusApproved = "approved"
	ReviewStatusRejected = "rejected"
)

// Review is a customer's rating/comment on a product they've bought.
// OrderID proves the purchase (see ReviewService.Create — only a delivered
// order containing this product qualifies), not just any authenticated
// user. The unique (product_id, user_id) index means one review per user
// per product, no matter how many times they bought it.
type Review struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ProductID uint      `json:"product_id"`
	UserID    uint      `json:"user_id"`
	OrderID   uint      `json:"order_id"`
	Rating    int       `json:"rating"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Product *Product `json:"product,omitempty"`
}
