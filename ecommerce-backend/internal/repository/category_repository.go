package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// ErrCategorySlugTaken is returned by Create when the slug unique
// constraint is violated — same detect-via-DB-error pattern as
// ProductRepository's ErrSlugTaken, for the same race-safety reason.
var ErrCategorySlugTaken = errors.New("category slug already taken")

// ErrCategoryInUse is returned by Delete when at least one product still
// references the category (products.category_id is ON DELETE RESTRICT).
var ErrCategoryInUse = errors.New("category is referenced by existing products")

const postgresForeignKeyViolation = "23503"

// CategoryRepository persists and retrieves Category records.
type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.Category, error)
	List(ctx context.Context) ([]model.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository builds a CategoryRepository backed by db.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *model.Category) error {
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
			return fmt.Errorf("repository: create category: %w", ErrCategorySlugTaken)
		}
		return fmt.Errorf("repository: create category: %w", err)
	}
	return nil
}

func (r *categoryRepository) Update(ctx context.Context, category *model.Category) error {
	if err := r.db.WithContext(ctx).Save(category).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
			return fmt.Errorf("repository: update category: %w", ErrCategorySlugTaken)
		}
		return fmt.Errorf("repository: update category: %w", err)
	}
	return nil
}

func (r *categoryRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Category{}, id)
	if err := result.Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresForeignKeyViolation {
			return fmt.Errorf("repository: delete category: %w", ErrCategoryInUse)
		}
		return fmt.Errorf("repository: delete category: %w", err)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("repository: delete category: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *categoryRepository) FindByID(ctx context.Context, id uint) (*model.Category, error) {
	var category model.Category
	err := r.db.WithContext(ctx).First(&category, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find category by id: %w", err)
	}
	return &category, nil
}

func (r *categoryRepository) List(ctx context.Context) ([]model.Category, error) {
	var categories []model.Category
	if err := r.db.WithContext(ctx).Order("name ASC").Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("repository: list categories: %w", err)
	}
	return categories, nil
}
