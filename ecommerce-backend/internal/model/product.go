package model

import (
	"time"

	"gorm.io/gorm"
)

// Product is the catalog entity. Price uses float64 rather than a decimal
// type — acceptable at this project's scale (IDR, no sub-cent amounts); a
// dedicated money type would be the right call for a system handling
// multi-currency or fractional-cent pricing.
type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `json:"name"`
	Slug        string         `json:"slug"`
	SKU         string         `json:"sku"`
	Description string         `json:"description"`
	Price       float64        `json:"price"`
	Stock       int            `json:"stock"`
	CategoryID  uint           `json:"category_id"`
	ImageURL    string         `json:"image_url"`
	CreatedAt   time.Time      `json:"created_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	Category *Category `json:"category,omitempty"`
}
