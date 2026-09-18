package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// RoleRepository retrieves Role records, primarily so services can assign
// default roles (e.g. "customer") without hardcoding role IDs.
type RoleRepository interface {
	FindByName(ctx context.Context, name string) (*model.Role, error)
}

type roleRepository struct {
	db *gorm.DB
}

// NewRoleRepository builds a RoleRepository backed by db.
func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find role by name: %w", err)
	}
	return &role, nil
}
