package model

// AdjustInventoryRequest is the payload for POST /api/admin/inventory/adjustments.
// Quantity's meaning depends on Type: for "in"/"out" it's the delta to
// apply; for "correction" it's the new absolute stock value.
type AdjustInventoryRequest struct {
	ProductID uint   `json:"product_id" binding:"required"`
	Type      string `json:"type" binding:"required,oneof=in out correction"`
	Quantity  int    `json:"quantity" binding:"gte=0"`
	Reason    string `json:"reason" binding:"required,max=500"`
	Reference string `json:"reference" binding:"max=255"`
}

// InventoryMovementListQuery binds GET
// /api/admin/inventory/:productId/movements query params.
type InventoryMovementListQuery struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}
