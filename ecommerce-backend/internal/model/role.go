package model

// Role groups a set of permissions and is assigned to users via user_roles.
type Role struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Name string `json:"name"`

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

// Permission is a single `resource:action` grant (e.g. "product:create").
type Permission struct {
	ID   uint   `gorm:"primaryKey" json:"id"`
	Code string `json:"code"`
}
