package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"gorm.io/gorm"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// CategoryService implements category CRUD: slug generation with collision
// retry (same pattern as ProductService), and delete guarded by the
// products.category_id foreign key (a category still in use can't be
// deleted out from under its products).
type CategoryService struct {
	categoryRepo repository.CategoryRepository
	auditLogRepo repository.AuditLogRepository
}

// NewCategoryService builds a CategoryService with its dependencies.
func NewCategoryService(categoryRepo repository.CategoryRepository, auditLogRepo repository.AuditLogRepository) *CategoryService {
	return &CategoryService{categoryRepo: categoryRepo, auditLogRepo: auditLogRepo}
}

// Create persists a new category under a slug derived from its name,
// retrying with a numeric suffix on collision.
func (s *CategoryService) Create(ctx context.Context, actorID uint, req model.CreateCategoryRequest) (*model.Category, error) {
	baseSlug := generateSlug(req.Name)
	category := &model.Category{Name: req.Name}

	for attempt := 0; attempt < maxSlugAttempts; attempt++ {
		category.Slug = baseSlug
		if attempt > 0 {
			category.Slug = fmt.Sprintf("%s-%d", baseSlug, attempt+1)
		}

		err := s.categoryRepo.Create(ctx, category)
		if err == nil {
			s.recordAudit(ctx, actorID, "category.create", category.ID, map[string]any{"name": category.Name})
			return category, nil
		}
		if errors.Is(err, repository.ErrCategorySlugTaken) {
			continue
		}
		return nil, apperror.Internal(fmt.Errorf("service: create category: %w", err))
	}

	return nil, apperror.Internal(fmt.Errorf("service: create category: no unique slug found after %d attempts for %q", maxSlugAttempts, baseSlug))
}

// Update renames a category. The slug is regenerated from the new name —
// unlike products, categories have no public detail page whose URL would
// break, so there's no reason to keep the old slug stable.
func (s *CategoryService) Update(ctx context.Context, actorID uint, id uint, req model.UpdateCategoryRequest) (*model.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update category: %w", err))
	}
	if category == nil {
		return nil, apperror.NotFound("Kategori tidak ditemukan", nil)
	}

	category.Name = req.Name
	category.Slug = generateSlug(req.Name)

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		if errors.Is(err, repository.ErrCategorySlugTaken) {
			return nil, apperror.DuplicateEntry("Nama kategori menghasilkan slug yang sudah dipakai kategori lain", err)
		}
		return nil, apperror.Internal(fmt.Errorf("service: update category: %w", err))
	}
	s.recordAudit(ctx, actorID, "category.update", category.ID, map[string]any{"name": category.Name})
	return category, nil
}

// Delete removes a category, rejecting with a conflict if any product still
// references it (see repository.ErrCategoryInUse).
func (s *CategoryService) Delete(ctx context.Context, actorID uint, id uint) error {
	if err := s.categoryRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, repository.ErrCategoryInUse) {
			return apperror.Conflict("Kategori masih dipakai oleh produk yang ada, tidak bisa dihapus", err)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Kategori tidak ditemukan", nil)
		}
		return apperror.Internal(fmt.Errorf("service: delete category: %w", err))
	}
	s.recordAudit(ctx, actorID, "category.delete", id, map[string]any{})
	return nil
}

// List returns every category, unpaginated — the catalog is small enough
// (portfolio scale) that a full list is what every caller (admin dropdown,
// public filter) actually wants.
func (s *CategoryService) List(ctx context.Context) ([]model.Category, error) {
	categories, err := s.categoryRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: list categories: %w", err))
	}
	return categories, nil
}

func (s *CategoryService) recordAudit(ctx context.Context, actorID uint, action string, resourceID uint, metadata map[string]any) {
	payload, err := json.Marshal(metadata)
	if err != nil {
		slog.Error("audit log: marshal metadata failed", slog.Any("err", err), slog.String("action", action))
		return
	}
	entry := &model.AuditLog{
		ActorID:    actorID,
		Action:     action,
		Resource:   "category",
		ResourceID: resourceID,
		Metadata:   string(payload),
	}
	if err := s.auditLogRepo.Create(ctx, entry); err != nil {
		slog.Error("audit log: write failed", slog.Any("err", err), slog.String("action", action))
	}
}
