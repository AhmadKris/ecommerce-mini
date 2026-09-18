package model

import "time"

// Inventory movement types. In/Out are deltas applied to the current stock
// (e.g. restock, damage/loss). Correction sets stock to an absolute target
// value instead — for reconciling against a physical stock count ("stock
// opname"), where the admin knows the true count, not the delta from it.
const (
	InventoryMovementTypeIn         = "in"
	InventoryMovementTypeOut        = "out"
	InventoryMovementTypeCorrection = "correction"
)

// InventoryMovement is an append-only record of a stock change — separate
// from audit_logs (which covers create/update/delete across every
// resource) because this one needs before/after quantities specifically,
// queryable per product, for a stock history view. Never updated or
// deleted once written.
type InventoryMovement struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ProductID      uint      `json:"product_id"`
	Type           string    `json:"type"`
	Quantity       int       `json:"quantity"`
	BeforeQuantity int       `json:"before_quantity"`
	AfterQuantity  int       `json:"after_quantity"`
	Reason         string    `json:"reason"`
	Reference      string    `json:"reference"`
	CreatedBy      uint      `json:"created_by"`
	CreatedAt      time.Time `json:"created_at"`

	Product *Product `json:"product,omitempty"`
}
