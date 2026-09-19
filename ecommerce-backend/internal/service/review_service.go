package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// ReviewService implements the review lifecycle: a customer can only review
// a product they've actually received (see Create), a review starts hidden
// until an admin moderates it (see UpdateStatus), and the public product
// page only ever sees approved ones (see ListApprovedForProduct).
type ReviewService struct {
	reviewRepo   repository.ReviewRepository
	productRepo  repository.ProductRepository
	auditLogRepo repository.AuditLogRepository
}

// NewReviewService builds a ReviewService with its dependencies.
func NewReviewService(reviewRepo repository.ReviewRepository, productRepo repository.ProductRepository, auditLogRepo repository.AuditLogRepository) *ReviewService {
	return &ReviewService{reviewRepo: reviewRepo, productRepo: productRepo, auditLogRepo: auditLogRepo}
}

// Create adds a review for the product identified by slug, on behalf of
// userID. The caller must have a delivered order containing this product —
// that's the purchase proof a review requires — and can only review a given
// product once (enforced again at the DB via a unique index, in case of a
// race between the check and the insert).
func (s *ReviewService) Create(ctx context.Context, userID uint, slug string, req model.CreateReviewRequest) (*model.Review, error) {
	product, err := s.productRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: create review: find product: %w", err))
	}
	if product == nil {
		return nil, apperror.NotFound("Produk tidak ditemukan", nil)
	}

	orderID, err := s.reviewRepo.FindDeliveredOrderID(ctx, userID, product.ID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: create review: find delivered order: %w", err))
	}
	if orderID == 0 {
		return nil, apperror.Forbidden("Kamu hanya bisa memberi ulasan untuk produk yang sudah diterima")
	}

	review := &model.Review{
		ProductID: product.ID,
		UserID:    userID,
		OrderID:   orderID,
		Rating:    req.Rating,
		Title:     req.Title,
		Body:      req.Body,
		Status:    model.ReviewStatusPending,
	}
	if err := s.reviewRepo.Create(ctx, review); err != nil {
		if errors.Is(err, repository.ErrReviewAlreadyExists) {
			return nil, apperror.DuplicateEntry("Kamu sudah memberi ulasan untuk produk ini", err)
		}
		return nil, apperror.Internal(fmt.Errorf("service: create review: %w", err))
	}

	s.recordAudit(ctx, userID, "review.create", review.ID, map[string]any{"product_id": product.ID, "rating": review.Rating})
	return review, nil
}

// ListApprovedForProduct returns the paginated, approved reviews for a
// product's public page.
func (s *ReviewService) ListApprovedForProduct(ctx context.Context, slug string, query model.ReviewListQuery) ([]model.Review, model.Meta, error) {
	product, err := s.productRepo.FindBySlug(ctx, slug)
	if err != nil {
		return nil, model.Meta{}, apperror.Internal(fmt.Errorf("service: list reviews: find product: %w", err))
	}
	if product == nil {
		return nil, model.Meta{}, apperror.NotFound("Produk tidak ditemukan", nil)
	}

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

	reviews, total, err := s.reviewRepo.ListByProductID(ctx, product.ID, model.ReviewStatusApproved, page, limit)
	if err != nil {
		return nil, model.Meta{}, apperror.Internal(fmt.Errorf("service: list reviews: %w", err))
	}
	return reviews, model.NewMeta(page, limit, total), nil
}

// ListAll returns every review regardless of status, for admin moderation.
func (s *ReviewService) ListAll(ctx context.Context, query model.ReviewListQuery) ([]model.Review, model.Meta, error) {
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

	reviews, total, err := s.reviewRepo.ListAll(ctx, page, limit)
	if err != nil {
		return nil, model.Meta{}, apperror.Internal(fmt.Errorf("service: list all reviews: %w", err))
	}
	return reviews, model.NewMeta(page, limit, total), nil
}

// UpdateStatus moderates a review — approving makes it visible on the
// product page, rejecting keeps it hidden permanently.
func (s *ReviewService) UpdateStatus(ctx context.Context, actorID uint, id uint, status string) (*model.Review, error) {
	existing, err := s.reviewRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update review status: %w", err))
	}
	if existing == nil {
		return nil, apperror.NotFound("Ulasan tidak ditemukan", nil)
	}

	review, err := s.reviewRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update review status: %w", err))
	}

	s.recordAudit(ctx, actorID, "review.status_change", id, map[string]any{"from": existing.Status, "to": status})
	return review, nil
}

func (s *ReviewService) recordAudit(ctx context.Context, actorID uint, action string, resourceID uint, metadata map[string]any) {
	payload, err := json.Marshal(metadata)
	if err != nil {
		slog.Error("audit log: marshal metadata failed", slog.Any("err", err), slog.String("action", action))
		return
	}
	entry := &model.AuditLog{
		ActorID:    actorID,
		Action:     action,
		Resource:   "review",
		ResourceID: resourceID,
		Metadata:   string(payload),
	}
	if err := s.auditLogRepo.Create(ctx, entry); err != nil {
		slog.Error("audit log: write failed", slog.Any("err", err), slog.String("action", action))
	}
}
