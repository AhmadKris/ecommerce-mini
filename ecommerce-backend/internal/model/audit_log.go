package model

import "time"

// AuditLog is an immutable record of a sensitive action — who did what to
// which resource, and when. Metadata is a JSON-encoded string (not a
// structured column) since its shape differs per action (product update logs
// changed fields, checkout logs order totals, etc.) — parse it only when
// actually inspecting a specific log, not on every read.
type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ActorID    uint      `json:"actor_id"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID uint      `json:"resource_id"`
	Metadata   string    `json:"metadata"`
	CreatedAt  time.Time `json:"created_at"`
}
