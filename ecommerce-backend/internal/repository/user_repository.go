// Package repository implements data access for each resource behind an
// interface defined alongside its implementation. A nil, nil return means
// "not found" — callers never match on gorm.ErrRecordNotFound directly.
package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// ErrEmailTaken is returned by Create when the email unique constraint is
// violated. Detected from the Postgres error code rather than a pre-check,
// so it also catches the race between two concurrent registrations for the
// same email.
var ErrEmailTaken = errors.New("email already taken")

const postgresUniqueViolation = "23505"

// UserRepository persists and retrieves User accounts.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	UpdatePassword(ctx context.Context, userID uint, passwordHash string) error
	// List returns a page of users (with roles preloaded), newest first,
	// optionally filtered by search matching name or email — backs the
	// admin customer list.
	List(ctx context.Context, search string, page, limit int) ([]model.User, int64, error)
	// UpdateRoles replaces the full role set for userID in a single
	// GORM association call (which runs in its own transaction). Passing an
	// empty slice removes all roles; callers must validate that at least one
	// role exists before calling.
	UpdateRoles(ctx context.Context, userID uint, roles []model.Role) error
}

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository builds a UserRepository backed by db.
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
			return fmt.Errorf("repository: create user: %w", ErrEmailTaken)
		}
		return fmt.Errorf("repository: create user: %w", err)
	}
	return nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		Where("email = ?", email).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find user by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find user by id: %w", err)
	}
	return &user, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uint, passwordHash string) error {
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
	if err != nil {
		return fmt.Errorf("repository: update password: %w", err)
	}
	return nil
}

func (r *userRepository) List(ctx context.Context, search string, page, limit int) ([]model.User, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.User{})
	if search != "" {
		like := "%" + search + "%"
		query = query.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repository: count users: %w", err)
	}

	listQuery := r.db.WithContext(ctx).Preload("Roles")
	if search != "" {
		like := "%" + search + "%"
		listQuery = listQuery.Where("name ILIKE ? OR email ILIKE ?", like, like)
	}

	var users []model.User
	err := listQuery.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list users: %w", err)
	}

	return users, total, nil
}

func (r *userRepository) UpdateRoles(ctx context.Context, userID uint, roles []model.Role) error {
	user := model.User{ID: userID}
	if err := r.db.WithContext(ctx).Model(&user).Association("Roles").Replace(roles); err != nil {
		return fmt.Errorf("repository: update user roles: %w", err)
	}
	return nil
}
