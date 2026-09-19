package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"gorm.io/gorm"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// PromotionService implements discount-code CRUD. Redeeming a code at
// checkout is OrderRepository.Checkout's job, not this service's — that
// path needs to run inside the same transaction as the stock decrement
// (see lockAndRedeemPromotion), which a separate service call could never
// guarantee.
type PromotionService struct {
	promotionRepo repository.PromotionRepository
	auditLogRepo  repository.AuditLogRepository
}

// NewPromotionService builds a PromotionService with its dependencies.
func NewPromotionService(promotionRepo repository.PromotionRepository, auditLogRepo repository.AuditLogRepository) *PromotionService {
	return &PromotionService{promotionRepo: promotionRepo, auditLogRepo: auditLogRepo}
}

// Create persists a new promotion code, normalized to uppercase (matches
// the repository's lookup normalization at redemption time).
func (s *PromotionService) Create(ctx context.Context, actorID uint, req model.CreatePromotionRequest) (*model.Promotion, error) {
	promotion := &model.Promotion{
		Code:            strings.ToUpper(req.Code),
		Type:            req.Type,
		Value:           req.Value,
		MinimumPurchase: req.MinimumPurchase,
		UsageLimit:      req.UsageLimit,
		StartsAt:        req.StartsAt,
		EndsAt:          req.EndsAt,
		Status:          model.PromotionStatusActive,
	}
	if err := s.promotionRepo.Create(ctx, promotion); err != nil {
		if errors.Is(err, repository.ErrPromotionCodeTaken) {
			return nil, apperror.DuplicateEntry("Kode promo sudah dipakai", err)
		}
		return nil, apperror.Internal(fmt.Errorf("service: create promotion: %w", err))
	}
	s.recordAudit(ctx, actorID, "promotion.create", promotion.ID, map[string]any{"code": promotion.Code})
	return promotion, nil
}

// Update replaces a promotion's terms. Code is intentionally not editable
// here — a code customers may have already seen/shared shouldn't silently
// start pointing at different terms; delete and recreate instead.
func (s *PromotionService) Update(ctx context.Context, actorID uint, id uint, req model.UpdatePromotionRequest) (*model.Promotion, error) {
	promotion, err := s.promotionRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update promotion: %w", err))
	}
	if promotion == nil {
		return nil, apperror.NotFound("Promo tidak ditemukan", nil)
	}

	promotion.Type = req.Type
	promotion.Value = req.Value
	promotion.MinimumPurchase = req.MinimumPurchase
	promotion.UsageLimit = req.UsageLimit
	promotion.StartsAt = req.StartsAt
	promotion.EndsAt = req.EndsAt
	promotion.Status = req.Status

	if err := s.promotionRepo.Update(ctx, promotion); err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update promotion: %w", err))
	}
	s.recordAudit(ctx, actorID, "promotion.update", promotion.ID, map[string]any{"status": promotion.Status})
	return promotion, nil
}

// Delete removes a promotion. Orders that already redeemed it keep their
// promotion_id (ON DELETE SET NULL, see migration) — deleting the
// promotion definition doesn't rewrite historical order discounts.
func (s *PromotionService) Delete(ctx context.Context, actorID uint, id uint) error {
	if err := s.promotionRepo.Delete(ctx, id); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperror.NotFound("Promo tidak ditemukan", nil)
		}
		return apperror.Internal(fmt.Errorf("service: delete promotion: %w", err))
	}
	s.recordAudit(ctx, actorID, "promotion.delete", id, map[string]any{})
	return nil
}

// List returns every promotion, unpaginated — same portfolio-scale
// reasoning as CategoryService.List.
func (s *PromotionService) List(ctx context.Context) ([]model.Promotion, error) {
	promotions, err := s.promotionRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: list promotions: %w", err))
	}
	return promotions, nil
}

func (s *PromotionService) recordAudit(ctx context.Context, actorID uint, action string, resourceID uint, metadata map[string]any) {
	payload, err := json.Marshal(metadata)
	if err != nil {
		slog.Error("audit log: marshal metadata failed", slog.Any("err", err), slog.String("action", action))
		return
	}
	entry := &model.AuditLog{
		ActorID:    actorID,
		Action:     action,
		Resource:   "promotion",
		ResourceID: resourceID,
		Metadata:   string(payload),
	}
	if err := s.auditLogRepo.Create(ctx, entry); err != nil {
		slog.Error("audit log: write failed", slog.Any("err", err), slog.String("action", action))
	}
}
