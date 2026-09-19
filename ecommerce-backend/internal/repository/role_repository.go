package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// RoleRepository retrieves Role records, primarily so services can assign
// default roles (e.g. "customer") without hardcoding role IDs, and so the
// admin role management endpoints can list and validate available roles.
type RoleRepository interface {
	FindByName(ctx context.Context, name string) (*model.Role, error)
	// List returns all roles with their permissions preloaded — used by the
	// admin role management UI to populate the available roles dropdown.
	List(ctx context.Context) ([]model.Role, error)
	// FindByNames fetches roles matching the given names in a single query.
	// Used to validate that every name in an UpdateUserRolesRequest refers to
	// an actual role before writing to user_roles. Names not found in the DB
	// won't appear in the result; the caller detects the mismatch by length.
	FindByNames(ctx context.Context, names []string) ([]model.Role, error)
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

func (r *roleRepository) List(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Preload("Permissions").Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("repository: list roles: %w", err)
	}
	return roles, nil
}

func (r *roleRepository) FindByNames(ctx context.Context, names []string) ([]model.Role, error) {
	var roles []model.Role
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&roles).Error
	if err != nil {
		return nil, fmt.Errorf("repository: find roles by names: %w", err)
	}
	return roles, nil
}
