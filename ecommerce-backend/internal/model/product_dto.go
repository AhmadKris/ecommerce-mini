package model

// CreateProductRequest is the payload for POST /api/admin/products. Price
// and Stock intentionally skip `required` — that validator tag treats a
// numeric zero as "missing", which would wrongly reject a legitimate
// zero-stock product.
type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=255"`
	Description string  `json:"description" binding:"max=5000"`
	Price       float64 `json:"price" binding:"gte=0"`
	Stock       int     `json:"stock" binding:"gte=0"`
	CategoryID  uint    `json:"category_id" binding:"required"`
	ImageURL    string  `json:"image_url" binding:"omitempty,url,max=500"`
}

// UpdateProductRequest is the payload for PUT /api/admin/products/:id.
// Every field is a pointer so the service can distinguish "not sent, leave
// unchanged" from "sent as its zero value" — e.g. updating Stock to 0 must
// take effect, not be treated as absent.
type UpdateProductRequest struct {
	Name        *string  `json:"name" binding:"omitempty,min=2,max=255"`
	Description *string  `json:"description" binding:"omitempty,max=5000"`
	Price       *float64 `json:"price" binding:"omitempty,gte=0"`
	Stock       *int     `json:"stock" binding:"omitempty,gte=0"`
	CategoryID  *uint    `json:"category_id" binding:"omitempty"`
	ImageURL    *string  `json:"image_url" binding:"omitempty,url,max=500"`
}

// ProductListQuery binds GET /api/products query params. Page/Limit are
// normalized (defaults + bounds) by the service, not here.
type ProductListQuery struct {
	Category string `form:"category"`
	Search   string `form:"search"`
	Page     int    `form:"page"`
	Limit    int    `form:"limit"`
}
