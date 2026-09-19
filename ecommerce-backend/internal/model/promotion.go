package model

import "time"

// Promotion discount types. Percentage's Value is 1-100 (a percent off the
// cart subtotal); Fixed's Value is a flat IDR amount off. Both are capped
// at the subtotal by the service so a discount can never make a total
// negative.
const (
	PromotionTypePercentage = "percentage"
	PromotionTypeFixed      = "fixed"
)

const (
	PromotionStatusActive   = "active"
	PromotionStatusInactive = "inactive"
)

// Promotion is a discount code redeemable at checkout. UsedCount is
// incremented atomically inside the same transaction as the checkout that
// redeems it (see OrderRepository.Checkout) — never read-then-write outside
// a lock, or two concurrent checkouts could both slip in under UsageLimit.
type Promotion struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Code            string    `json:"code"`
	Type            string    `json:"type"`
	Value           float64   `json:"value"`
	MinimumPurchase float64   `json:"minimum_purchase"`
	UsageLimit      int       `json:"usage_limit"`
	UsedCount       int       `json:"used_count"`
	StartsAt        time.Time `json:"starts_at"`
	EndsAt          time.Time `json:"ends_at"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}
