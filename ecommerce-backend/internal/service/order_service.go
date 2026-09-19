package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// orderStatusTransitions lists, for each status, which statuses it's
// allowed to move to next. Shipped/Delivered are terminal for cancellation
// on purpose — once a courier has it, "cancelled" no longer reflects
// reality (that's a return/refund flow, a different feature). Delivered and
// Cancelled have no outgoing edges at all: both are end states.
var orderStatusTransitions = map[string][]string{
	model.OrderStatusPending:    {model.OrderStatusPaid, model.OrderStatusCancelled},
	model.OrderStatusPaid:       {model.OrderStatusProcessing, model.OrderStatusCancelled},
	model.OrderStatusProcessing: {model.OrderStatusShipped, model.OrderStatusCancelled},
	model.OrderStatusShipped:    {model.OrderStatusDelivered},
	model.OrderStatusDelivered:  {},
	model.OrderStatusCancelled:  {},
}

// OrderService implements checkout and order history.
type OrderService struct {
	orderRepo    repository.OrderRepository
	auditLogRepo repository.AuditLogRepository
}

// NewOrderService builds an OrderService with its dependencies.
func NewOrderService(orderRepo repository.OrderRepository, auditLogRepo repository.AuditLogRepository) *OrderService {
	return &OrderService{orderRepo: orderRepo, auditLogRepo: auditLogRepo}
}

// Checkout turns userID's cart into an order. All atomicity and locking
// live in the repository (see OrderRepository.Checkout) — this method's
// job is just translating its sentinel errors into client-facing ones.
func (s *OrderService) Checkout(ctx context.Context, userID uint, req model.CheckoutRequest) (*model.Order, error) {
	order, err := s.orderRepo.Checkout(ctx, userID, req.ShippingAddress, req.PromoCode)
	if err != nil {
		if errors.Is(err, repository.ErrEmptyCart) {
			return nil, apperror.Validation("Keranjang Anda kosong, tidak bisa checkout", nil)
		}
		if errors.Is(err, repository.ErrInsufficientStock) {
			return nil, apperror.Conflict("Stok tidak mencukupi untuk salah satu produk di keranjang", err)
		}
		if errors.Is(err, repository.ErrPromotionInvalid) {
			return nil, apperror.Validation("Kode promo tidak valid atau sudah tidak berlaku", nil)
		}
		if errors.Is(err, repository.ErrPromotionExhausted) {
			return nil, apperror.Conflict("Kode promo sudah mencapai batas penggunaan", err)
		}
		if errors.Is(err, repository.ErrPromotionMinimumNotMet) {
			return nil, apperror.Validation("Total belanja belum memenuhi syarat minimum kode promo ini", nil)
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

// ListAll returns a page of every order across all users, for the admin
// order list — same pagination normalization as List.
func (s *OrderService) ListAll(ctx context.Context, query model.OrderListQuery) ([]model.Order, model.Meta, error) {
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

	orders, total, err := s.orderRepo.ListAll(ctx, page, limit)
	if err != nil {
		return nil, model.Meta{}, apperror.Internal(fmt.Errorf("service: list all orders: %w", err))
	}

	return orders, model.NewMeta(page, limit, total), nil
}

// GetByID returns a single order with its items, for the admin order detail
// page — no ownership check, since only order:read_all-gated callers reach
// this (any admin may look up any order, unlike the customer-facing List).
func (s *OrderService) GetByID(ctx context.Context, orderID uint) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: get order: %w", err))
	}
	if order == nil {
		return nil, apperror.NotFound("Order tidak ditemukan", nil)
	}
	return order, nil
}

// UpdateStatus transitions an order to newStatus, rejecting any transition
// not in orderStatusTransitions (e.g. "delivered" back to "pending", or any
// move out of a terminal state) with a 409 rather than silently accepting
// it — an order's status history should only ever move forward.
func (s *OrderService) UpdateStatus(ctx context.Context, actorID uint, orderID uint, newStatus string) (*model.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update order status: %w", err))
	}
	if order == nil {
		return nil, apperror.NotFound("Order tidak ditemukan", nil)
	}

	allowed := orderStatusTransitions[order.Status]
	if !slices.Contains(allowed, newStatus) {
		return nil, apperror.Conflict(
			fmt.Sprintf("Order dengan status %q tidak bisa diubah ke %q", order.Status, newStatus),
			nil,
		)
	}

	previousStatus := order.Status
	updated, err := s.orderRepo.UpdateStatus(ctx, orderID, newStatus)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: update order status: %w", err))
	}

	s.recordStatusChangeAudit(ctx, actorID, orderID, previousStatus, newStatus)
	return updated, nil
}

func (s *OrderService) recordStatusChangeAudit(ctx context.Context, actorID uint, orderID uint, from, to string) {
	payload, err := json.Marshal(map[string]any{"from": from, "to": to})
	if err != nil {
		slog.Error("audit log: marshal metadata failed", slog.Any("err", err), slog.String("action", "order.status_change"))
		return
	}
	entry := &model.AuditLog{
		ActorID:    actorID,
		Action:     "order.status_change",
		Resource:   "order",
		ResourceID: orderID,
		Metadata:   string(payload),
	}
	if err := s.auditLogRepo.Create(ctx, entry); err != nil {
		slog.Error("audit log: write failed", slog.Any("err", err), slog.String("action", "order.status_change"))
	}
}
