package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"regexp"
	"strings"

	"gorm.io/gorm"

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
	auditLogRepo repository.AuditLogRepository
}

// NewProductService builds a ProductService with its dependencies.
func NewProductService(productRepo repository.ProductRepository, categoryRepo repository.CategoryRepository, auditLogRepo repository.AuditLogRepository) *ProductService {
	return &ProductService{productRepo: productRepo, categoryRepo: categoryRepo, auditLogRepo: auditLogRepo}
}

// Create validates the target category exists, then persists the product
// under a slug derived from its name — retrying with a numeric suffix if
// that slug is already taken by another (non-deleted) product.
func (s *ProductService) Create(ctx context.Context, actorID uint, req model.CreateProductRequest) (*model.Product, error) {
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
			s.recordAudit(ctx, actorID, "product.create", product.ID, map[string]any{
				"name": product.Name, "price": product.Price, "stock": product.Stock,
			})
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
func (s *ProductService) Update(ctx context.Context, actorID uint, id uint, req model.UpdateProductRequest) (*model.Product, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update product: %w", err))
	}
	if product == nil {
		return nil, apperror.NotFound("Produk tidak ditemukan", nil)
	}

	changes := map[string]any{}

	if req.CategoryID != nil {
		if err := s.assertCategoryExists(ctx, *req.CategoryID); err != nil {
			return nil, err
		}
		product.CategoryID = *req.CategoryID
		changes["category_id"] = *req.CategoryID
	}
	if req.Name != nil {
		product.Name = *req.Name
		changes["name"] = *req.Name
	}
	if req.Description != nil {
		product.Description = *req.Description
	}
	if req.Price != nil {
		product.Price = *req.Price
		changes["price"] = *req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
		changes["stock"] = *req.Stock
	}
	if req.ImageURL != nil {
		if *req.ImageURL != "" && !isValidURL(*req.ImageURL) {
			return nil, apperror.Validation("Data produk tidak valid", []string{"image_url: url"})
		}
		product.ImageURL = *req.ImageURL
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update product: %w", err))
	}
	s.recordAudit(ctx, actorID, "product.update", product.ID, changes)
	return product, nil
}

// Delete soft-deletes a product. It stays visible on existing orders/carts
// (see ProductRepository.Delete) but disappears from the public catalog and
// from admin's product list immediately.
func (s *ProductService) Delete(ctx context.Context, actorID uint, id uint) error {
	if err := s.productRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Produk tidak ditemukan", nil)
		}
		return apperror.Internal(fmt.Errorf("service: delete product: %w", err))
	}
	s.recordAudit(ctx, actorID, "product.delete", id, map[string]any{})
	return nil
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

// recordAudit is best-effort: a failed audit write logs the failure but
// never fails the product write that triggered it — the write already
// succeeded, and losing an audit entry is preferable to telling an admin
// their product update failed when it didn't.
func (s *ProductService) recordAudit(ctx context.Context, actorID uint, action string, resourceID uint, metadata map[string]any) {
	payload, err := json.Marshal(metadata)
	if err != nil {
		slog.Error("audit log: marshal metadata failed", slog.Any("err", err), slog.String("action", action))
		return
	}
	entry := &model.AuditLog{
		ActorID:    actorID,
		Action:     action,
		Resource:   "product",
		ResourceID: resourceID,
		Metadata:   string(payload),
	}
	if err := s.auditLogRepo.Create(ctx, entry); err != nil {
		slog.Error("audit log: write failed", slog.Any("err", err), slog.String("action", action))
	}
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

// isValidURL reports whether raw parses as an absolute URL with a scheme
// and host (e.g. "https://example.com/x.jpg") — the same shape the
// validator's `url` tag checks, applied manually here because that tag's
// `omitempty` doesn't skip a non-nil pointer to "" (see UpdateProductRequest).
func isValidURL(raw string) bool {
	parsed, err := url.ParseRequestURI(raw)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}
