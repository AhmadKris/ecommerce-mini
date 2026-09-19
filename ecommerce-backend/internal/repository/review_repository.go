package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

// ErrReviewAlreadyExists is returned by Create when the user already
// reviewed this product (unique (product_id, user_id) index).
var ErrReviewAlreadyExists = errors.New("user already reviewed this product")

// ReviewRepository persists product reviews and their moderation status.
type ReviewRepository interface {
	Create(ctx context.Context, review *model.Review) error
	FindByID(ctx context.Context, id uint) (*model.Review, error)
	UpdateStatus(ctx context.Context, id uint, status string) (*model.Review, error)
	ListByProductID(ctx context.Context, productID uint, status string, page, limit int) ([]model.Review, int64, error)
	ListAll(ctx context.Context, page, limit int) ([]model.Review, int64, error)
	// FindDeliveredOrderID returns the ID of a delivered order belonging to
	// userID that contains productID, or 0 if none exists — the purchase
	// proof a review requires. Zero, not an error, means "not eligible";
	// an error means the query itself failed.
	FindDeliveredOrderID(ctx context.Context, userID, productID uint) (uint, error)
}

type reviewRepository struct {
	db *gorm.DB
}

// NewReviewRepository builds a ReviewRepository backed by db.
func NewReviewRepository(db *gorm.DB) ReviewRepository {
	return &reviewRepository{db: db}
}

func (r *reviewRepository) Create(ctx context.Context, review *model.Review) error {
	if err := r.db.WithContext(ctx).Create(review).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == postgresUniqueViolation {
			return fmt.Errorf("repository: create review: %w", ErrReviewAlreadyExists)
		}
		return fmt.Errorf("repository: create review: %w", err)
	}
	return nil
}

func (r *reviewRepository) FindByID(ctx context.Context, id uint) (*model.Review, error) {
	var review model.Review
	err := r.db.WithContext(ctx).First(&review, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("repository: find review by id: %w", err)
	}
	return &review, nil
}

func (r *reviewRepository) UpdateStatus(ctx context.Context, id uint, status string) (*model.Review, error) {
	if err := r.db.WithContext(ctx).Model(&model.Review{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return nil, fmt.Errorf("repository: update review status: %w", err)
	}
	return r.FindByID(ctx, id)
}

func (r *reviewRepository) ListByProductID(ctx context.Context, productID uint, status string, page, limit int) ([]model.Review, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Review{}).Where("product_id = ?", productID)
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repository: count reviews: %w", err)
	}

	var reviews []model.Review
	listQuery := r.db.WithContext(ctx).Where("product_id = ?", productID)
	if status != "" {
		listQuery = listQuery.Where("status = ?", status)
	}
	err := listQuery.Order("created_at DESC").Offset((page - 1) * limit).Limit(limit).Find(&reviews).Error
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list reviews by product: %w", err)
	}

	return reviews, total, nil
}

func (r *reviewRepository) ListAll(ctx context.Context, page, limit int) ([]model.Review, int64, error) {
	var total int64
	if err := r.db.WithContext(ctx).Model(&model.Review{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repository: count all reviews: %w", err)
	}

	var reviews []model.Review
	err := r.db.WithContext(ctx).
		Preload("Product").
		Order("created_at DESC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&reviews).Error
	if err != nil {
		return nil, 0, fmt.Errorf("repository: list all reviews: %w", err)
	}

	return reviews, total, nil
}

func (r *reviewRepository) FindDeliveredOrderID(ctx context.Context, userID, productID uint) (uint, error) {
	var orderItem model.OrderItem
	err := r.db.WithContext(ctx).
		Joins("JOIN orders ON orders.id = order_items.order_id").
		Where("orders.user_id = ? AND order_items.product_id = ? AND orders.status = ?", userID, productID, model.OrderStatusDelivered).
		First(&orderItem).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("repository: find delivered order: %w", err)
	}
	return orderItem.OrderID, nil
}
