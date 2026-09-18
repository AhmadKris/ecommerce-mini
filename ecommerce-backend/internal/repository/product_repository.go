package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// ErrSlugTaken is returned by Create when the slug unique constraint is
// violated, letting the service retry with a suffixed slug instead of
// pre-checking (and racing) for availability.
var ErrSlugTaken = errors.New("product slug already taken")

// ProductFilter narrows ProductRepository.List. Page/Limit are expected to
// already be normalized by the service.
type ProductFilter struct {
	CategorySlug string
	Search       string
	Page         int
	Limit        int
}

// ProductRepository persists and retrieves Product catalog entries.
type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	Update(ctx context.Context, product *model.Product) error
	FindByID(ctx context.Context, id uint) (*model.Product, error)
	FindBySlug(ctx context.Context, slug string) (*model.Product, error)
	List(ctx context.Context, filter ProductFilter) ([]model.Product, int64, error)
}

type productRepository struct {
	db *gorm.DB
}

// NewProductRepository builds a ProductRepository backed by db.
func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
			return fmt.Errorf("repository: create product: %w", ErrSlugTaken)
		}
		return fmt.Errorf("repository: create product: %w", err)
	}
	return nil
}

func (r *productRepository) Update(ctx context.Context, product *model.Product) error {
	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		return fmt.Errorf("repository: update product: %w", err)
	}
	return nil
}

func (r *productRepository) FindByID(ctx context.Context, id uint) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Preload("Category").First(&product, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find product by id: %w", err)
	}
	return &product, nil
}

func (r *productRepository) FindBySlug(ctx context.Context, slug string) (*model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("slug = ?", slug).First(&product).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find product by slug: %w", err)
	}
	return &product, nil
}

func (r *productRepository) List(ctx context.Context, filter ProductFilter) ([]model.Product, int64, error) {
	var total int64
	if err := r.baseQuery(ctx, filter).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repository: count products: %w", err)
	}

	var products []model.Product
	err := r.baseQuery(ctx, filter).
		Preload("Category").
		Order("products.created_at DESC").
		Offset((filter.Page - 1) * filter.Limit).
		Limit(filter.Limit).
		Find(&products).Error
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list products: %w", err)
	}

	return products, total, nil
}

// baseQuery builds a fresh statement per call (rather than reusing one
// across Count and Find) so clauses from one call never leak into the other.
func (r *productRepository) baseQuery(ctx context.Context, filter ProductFilter) *gorm.DB {
	query := r.db.WithContext(ctx).Model(&model.Product{})
	if filter.CategorySlug != "" {
		query = query.Joins("JOIN categories ON categories.id = products.category_id").
			Where("categories.slug = ?", filter.CategorySlug)
	}
	if filter.Search != "" {
		query = query.Where("products.name ILIKE ?", "%"+filter.Search+"%")
	}
	return query
}
