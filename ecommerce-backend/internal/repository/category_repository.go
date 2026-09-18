package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// CategoryRepository retrieves Category records. Category CRUD itself isn't
// part of Fase 1 — this exists so ProductService can validate a
// category_id before creating/updating a product.
type CategoryRepository interface {
	FindByID(ctx context.Context, id uint) (*model.Category, error)
}

type categoryRepository struct {
	db *gorm.DB
}

// NewCategoryRepository builds a CategoryRepository backed by db.
func NewCategoryRepository(db *gorm.DB) CategoryRepository {
	return &categoryRepository{db: db}
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
