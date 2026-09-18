package service

import (
	"context"
	"errors"
	"fmt"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// OrderService implements checkout and order history.
type OrderService struct {
	orderRepo repository.OrderRepository
}

// NewOrderService builds an OrderService with its dependencies.
func NewOrderService(orderRepo repository.OrderRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo}
}

// Checkout turns userID's cart into an order. All atomicity and locking
// live in the repository (see OrderRepository.Checkout) — this method's
// job is just translating its sentinel errors into client-facing ones.
func (s *OrderService) Checkout(ctx context.Context, userID uint, req model.CheckoutRequest) (*model.Order, error) {
	order, err := s.orderRepo.Checkout(ctx, userID, req.ShippingAddress)
	if err != nil {
		if errors.Is(err, repository.ErrEmptyCart) {
			return nil, apperror.Validation("Keranjang Anda kosong, tidak bisa checkout", nil)
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			return nil, apperror.Conflict("Stok tidak mencukupi untuk salah satu produk di keranjang", err)
		}
		return nil, apperror.Internal(fmt.Errorf("service: checkout: %w", err))
	}
	return order, nil
}

// List returns a page of userID's own order history, newest first.
func (s *OrderService) List(ctx context.Context, userID uint, query model.OrderListQuery) ([]model.Order, model.Meta, error) {
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

	orders, total, err := s.orderRepo.ListByUserID(ctx, userID, page, limit)
	if err != nil {
		return nil, model.Meta{}, apperror.Internal(fmt.Errorf("service: list orders: %w", err))
	}

	return orders, model.NewMeta(page, limit, total), nil
}
