package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"ecommerce-backend/internal/model"
)

// ErrPromotionCodeTaken is returned by Create when the code unique
// constraint is violated — same detect-via-DB-error pattern as
// ErrSlugTaken/ErrCategorySlugTaken.
var ErrPromotionCodeTaken = errors.New("promotion code already taken")

// ErrPromotionInvalid covers "doesn't exist", "inactive", and "outside its
// active date range" — all three mean the same thing to a checkout caller
// (this code can't be redeemed right now), so they share one sentinel
// rather than forcing the service to distinguish reasons a customer can't
// act on differently anyway.
var ErrPromotionInvalid = errors.New("promotion code invalid or not active")

// ErrPromotionMinimumNotMet means the cart subtotal is below the
// promotion's minimum_purchase.
var ErrPromotionMinimumNotMet = errors.New("cart subtotal below promotion minimum purchase")

// ErrPromotionExhausted means the promotion's usage_limit has already been
// reached.
var ErrPromotionExhausted = errors.New("promotion usage limit reached")

// PromotionRepository persists discount codes.
type PromotionRepository interface {
	Create(ctx context.Context, promotion *model.Promotion) error
	Update(ctx context.Context, promotion *model.Promotion) error
	Delete(ctx context.Context, id uint) error
	FindByID(ctx context.Context, id uint) (*model.Promotion, error)
	List(ctx context.Context) ([]model.Promotion, error)
}

type promotionRepository struct {
	db *gorm.DB
}

// NewPromotionRepository builds a PromotionRepository backed by db.
func NewPromotionRepository(db *gorm.DB) PromotionRepository {
	return &promotionRepository{db: db}
}

func (r *promotionRepository) Create(ctx context.Context, promotion *model.Promotion) error {
	promotion.Code = strings.ToUpper(promotion.Code)
	if err := r.db.WithContext(ctx).Create(promotion).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
			return fmt.Errorf("repository: create promotion: %w", ErrPromotionCodeTaken)
		}
		return fmt.Errorf("repository: create promotion: %w", err)
	}
	return nil
}

func (r *promotionRepository) Update(ctx context.Context, promotion *model.Promotion) error {
	if err := r.db.WithContext(ctx).Save(promotion).Error; err != nil {
		return fmt.Errorf("repository: update promotion: %w", err)
	}
	return nil
}

func (r *promotionRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Promotion{}, id)
	if result.Error != nil {
		return fmt.Errorf("repository: delete promotion: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("repository: delete promotion: %w", gorm.ErrRecordNotFound)
	}
	return nil
}

func (r *promotionRepository) FindByID(ctx context.Context, id uint) (*model.Promotion, error) {
	var promotion model.Promotion
	err := r.db.WithContext(ctx).First(&promotion, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find promotion by id: %w", err)
	}
	return &promotion, nil
}

func (r *promotionRepository) List(ctx context.Context) ([]model.Promotion, error) {
	var promotions []model.Promotion
	if err := r.db.WithContext(ctx).Order("created_at DESC").Find(&promotions).Error; err != nil {
		return nil, fmt.Errorf("repository: list promotions: %w", err)
	}
	return promotions, nil
}

// lockAndRedeemPromotion locks the promotion row (SELECT ... FOR UPDATE),
// validates it against subtotal and the current time, and — only if
// valid — increments used_count in the same statement's transaction,
// returning the discount to apply. Called from OrderRepository.Checkout
// with its own tx, the same row-lock-then-write pattern as
// lockAndReserveStock: without the lock, two concurrent checkouts redeeming
// the last use of a usage_limit=1 code could both read used_count below the
// limit and both succeed, over-redeeming it exactly like unlocked stock
// reads would oversell.
func lockAndRedeemPromotion(tx *gorm.DB, code string, subtotal float64) (*model.Promotion, float64, error) {
	var promotion model.Promotion
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("code = ?", strings.ToUpper(code)).
		First(&promotion).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, 0, ErrPromotionInvalid
	}
	if err != nil {
		return nil, 0, err
	}

	now := time.Now()
	if promotion.Status != model.PromotionStatusActive || now.Before(promotion.StartsAt) || now.After(promotion.EndsAt) {
		return nil, 0, ErrPromotionInvalid
	}
	if promotion.UsedCount >= promotion.UsageLimit {
		return nil, 0, ErrPromotionExhausted
	}
	if subtotal < promotion.MinimumPurchase {
		return nil, 0, ErrPromotionMinimumNotMet
	}

	discount := promotion.Value
	if promotion.Type == model.PromotionTypePercentage {
		discount = subtotal * promotion.Value / 100
	}
	if discount > subtotal {
		discount = subtotal
	}

	promotion.UsedCount++
	if err := tx.Save(&promotion).Error; err != nil {
		return nil, 0, err
	}

	return &promotion, discount, nil
}
