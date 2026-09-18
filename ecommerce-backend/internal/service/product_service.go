package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

const (
	defaultPage  = 1
	defaultLimit = 10
	maxLimit     = 100

	maxSlugAttempts = 5
)

var slugInvalidChars = regexp.MustCompile(`[^a-z0-9]+`)

// ProductService implements the product catalog's business rules: category
// existence on write, slug generation with collision retry, and pagination
// normalization on read.
type ProductService struct {
	productRepo  repository.ProductRepository
	categoryRepo repository.CategoryRepository
}

// NewProductService builds a ProductService with its dependencies.
func NewProductService(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository) *ProductService {
	return &ProductService{productRepo: productRepo, categoryRepo: categoryRepo}
}

// Create validates the target category exists, then persists the product
// under a slug derived from its name — retrying with a numeric suffix if
// that slug is already taken by another (non-deleted) product.
func (s *ProductService) Create(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	if err := s.assertCategoryExists(ctx, req.CategoryID); err != nil {
		return nil, err
	}

	baseSlug := generateSlug(req.Name)
	product := &model.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
		CategoryID:  req.CategoryID,
		ImageURL:    req.ImageURL,
	}

	for attempt := 0; attempt < maxSlugAttempts; attempt++ {
		product.Slug = baseSlug
		if attempt > 0 {
			product.Slug = fmt.Sprintf("%s-%d", baseSlug, attempt+1)
		}

		err := s.productRepo.Create(ctx, product)
		if err == nil {
			return product, nil
		}
		if errors.Is(err, repository.ErrSlugTaken) {
			continue
		}
		return nil, apperror.Internal(fmt.Errorf("service: create product: %w", err))
	}

	return nil, apperror.Internal(fmt.Errorf("service: create product: no unique slug found after %d attempts for %q", maxSlugAttempts, baseSlug))
}

// Update applies only the fields present in req (see UpdateProductRequest's
// pointer fields) to the existing product. The slug is never changed here —
// it's assigned once at creation so published product URLs stay stable.
func (s *ProductService) Update(ctx context.Context, id uint, req model.UpdateProductRequest) (*model.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update product: %w", err))
	}
	if product == nil {
		return nil, apperror.NotFound("Produk tidak ditemukan", nil)
	}

	if req.CategoryID != nil {
		if err := s.assertCategoryExists(ctx, *req.CategoryID); err != nil {
			return nil, err
		}
		product.CategoryID = *req.CategoryID
	}
	if req.Name != nil {
		product.Name = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}
	if req.ImageURL != nil {
		product.ImageURL = *req.ImageURL
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update product: %w", err))
	}
	return product, nil
}

// GetBySlug returns a single product for the public product detail page.
func (s *ProductService) GetBySlug(ctx context.Context, slug string) (*model.Product, error) {
	product, err := s.productRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: get product by slug: %w", err))
	}
	if product == nil {
		return nil, apperror.NotFound("Produk tidak ditemukan", nil)
	}
	return product, nil
}

// List returns a page of products matching query, normalizing page/limit to
// safe bounds (page >= 1, 1 <= limit <= 100).
func (s *ProductService) List(ctx context.Context, query model.ProductListQuery) ([]model.Product, model.Meta, error) {
	page := query.Page
	if page < 1 {
		page = defaultPage
	}
	limit := query.Limit
	if limit < 1 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}

	products, total, err := s.productRepo.List(ctx, repository.ProductFilter{
		CategorySlug: query.Category,
		Search:       query.Search,
		Page:         page,
		Limit:        limit,
	})
	if err != nil {
		return nil, model.Meta{}, apperror.Internal(fmt.Errorf("service: list products: %w", err))
	}

	return products, model.NewMeta(page, limit, total), nil
}

func (s *ProductService) assertCategoryExists(ctx context.Context, categoryID uint) error {
	category, err := s.categoryRepo.FindByID(ctx, categoryID)
	if err != nil {
		return apperror.Internal(fmt.Errorf("service: check category: %w", err))
	}
	if category == nil {
		return apperror.Validation("Kategori tidak ditemukan", []string{"category_id: does not exist"})
	}
	return nil
}

// generateSlug lowercases name and replaces runs of non-alphanumeric
// characters with a single hyphen (e.g. "Kopi Susu Gula Aren!" ->
// "kopi-susu-gula-aren").
func generateSlug(name string) string {
	slug := slugInvalidChars.ReplaceAllString(strings.ToLower(name), "-")
	return strings.Trim(slug, "-")
}
