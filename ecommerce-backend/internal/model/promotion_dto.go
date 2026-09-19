package model

import "time"

// CreatePromotionRequest is the payload for POST /api/admin/promotions.
type CreatePromotionRequest struct {
	Code            string    `json:"code" binding:"required,min=3,max=50"`
	Type            string    `json:"type" binding:"required,oneof=percentage fixed"`
	Value           float64   `json:"value" binding:"required,gt=0"`
	MinimumPurchase float64   `json:"minimum_purchase" binding:"gte=0"`
	UsageLimit      int       `json:"usage_limit" binding:"required,gt=0"`
	StartsAt        time.Time `json:"starts_at" binding:"required"`
	EndsAt          time.Time `json:"ends_at" binding:"required,gtfield=StartsAt"`
}

// UpdatePromotionRequest is the payload for PUT /api/admin/promotions/:id —
// a full replace, same reasoning as UpdateAddressRequest (every field is
// cheap to resend, no explicit-zero ambiguity worth pointer fields here).
// UsedCount is deliberately not editable through this DTO — it only ever
// changes via a checkout redeeming the code.
type UpdatePromotionRequest struct {
	Type            string    `json:"type" binding:"required,oneof=percentage fixed"`
	Value           float64   `json:"value" binding:"required,gt=0"`
	MinimumPurchase float64   `json:"minimum_purchase" binding:"gte=0"`
	UsageLimit      int       `json:"usage_limit" binding:"required,gt=0"`
	StartsAt        time.Time `json:"starts_at" binding:"required"`
	EndsAt          time.Time `json:"ends_at" binding:"required,gtfield=StartsAt"`
	Status          string    `json:"status" binding:"required,oneof=active inactive"`
}
