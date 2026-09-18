// Package model holds entity structs and request/response DTOs for every
// resource (user, role, permission, product, ...). Entities carry gorm and
// json tags together; each request shape gets its own struct even when its
// fields currently mirror the entity.
package model

import (
	"time"

	"gorm.io/gorm"
)

// User is the account entity. PasswordHash is never serialized to JSON.
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `json:"name"`
	Email        string         `json:"email"`
	PasswordHash string         `json:"-"`
	CreatedAt    time.Time      `json:"created_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Roles []Role `gorm:"many2many:user_roles;" json:"roles,omitempty"`
}

// RoleNames returns the names of every role assigned to the user.
func (u *User) RoleNames() []string {
	names := make([]string, len(u.Roles))
	for i, role := range u.Roles {
		names[i] = role.Name
	}
	return names
}

// PermissionCodes returns the deduplicated permission codes granted to the
// user across all of its roles.
func (u *User) PermissionCodes() []string {
	seen := make(map[string]struct{})
	codes := make([]string, 0)
	for _, role := range u.Roles {
		for _, permission := range role.Permissions {
			if _, ok := seen[permission.Code]; ok {
				continue
			}
			seen[permission.Code] = struct{}{}
			codes = append(codes, permission.Code)
		}
	}
	return codes
}
