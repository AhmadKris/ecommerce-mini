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

// ErrSKUTaken is returned by Create/Update when the sku unique constraint is
// violated. Unlike the slug, the SKU is admin-chosen, not generated, so
// there's nothing to retry with — the caller surfaces this as a validation
// error asking for a different SKU.
var ErrSKUTaken = errors.New("product sku already taken")

// Sort values ProductFilter.Sort accepts — anything else falls back to
// SortNewest. Kept as a closed set (not a raw ORDER BY string) so a filter
// value can never become a SQL-injection vector.
const (
	SortNewest    = "newest"
	SortPriceAsc  = "price_asc"
	SortPriceDesc = "price_desc"
)

var productSortColumns = map[string]string{
	SortNewest:    "products.created_at DESC",
	SortPriceAsc:  "products.price ASC",
	SortPriceDesc: "products.price DESC",
}

// ProductFilter narrows ProductRepository.List. Page/Limit/Sort are
// expected to already be normalized by the service.
type ProductFilter struct {
	CategorySlug string
	Search       string
	Sort         string
	Page         int
	Limit        int
}

// ProductRepository persists and retrieves Product catalog entries.
type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id uint) error
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
		return fmt.Errorf("repository: create product: %w", translateUniqueViolation(err))
	}
	return nil
}

func (r *productRepository) Update(ctx context.Context, product *model.Product) error {
	if err := r.db.WithContext(ctx).Save(product).Error; err != nil {
		return fmt.Errorf("repository: update product: %w", translateUniqueViolation(err))
	}
	return nil
}

// translateUniqueViolation maps a Postgres unique-violation on products to
// its sentinel error by which partial unique index fired — ConstraintName is
// the index name for a unique index violation. Any other error, including a
// unique violation on an index this doesn't recognize, passes through
// unchanged.
func translateUniqueViolation(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != postgresUniqueViolation {
		return err
	}
	switch pgErr.ConstraintName {
	case "idx_products_sku_active":
		return ErrSKUTaken
	case "idx_products_slug_active":
		return ErrSlugTaken
	default:
		return err
	}
}

// Delete soft-deletes a product (sets deleted_at, per model.Product's
// gorm.DeletedAt) rather than a hard DELETE — order_items/cart_items still
// reference the row by product_id, and a hard delete would either violate
// that foreign key or destroy history. idx_products_slug_active is already
// scoped to deleted_at IS NULL, so the slug frees up for reuse immediately.
func (r *productRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Product{}, id)
	if result.Error != nil {
		return fmt.Errorf("repository: delete product: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("repository: delete product: %w", gorm.ErrRecordNotFound)
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

	orderBy, ok := productSortColumns[filter.Sort]
	if !ok {
		orderBy = productSortColumns[SortNewest]
	}

	var products []model.Product
	err := r.baseQuery(ctx, filter).
		Preload("Category").
		Order(orderBy).
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
		like := "%" + filter.Search + "%"
		query = query.Where("products.name ILIKE ? OR products.sku ILIKE ?", like, like)
	}
	return query
}
