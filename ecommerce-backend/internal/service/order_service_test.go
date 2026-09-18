package service

import (
	"context"
	"testing"

	"ecommerce-backend/internal/model"
)

type fakeOrderRepo struct {
	orders map[uint]*model.Order
}

func newFakeOrderRepo(orders ...*model.Order) *fakeOrderRepo {
	repo := &fakeOrderRepo{orders: map[uint]*model.Order{}}
	for _, order := range orders {
		repo.orders[order.ID] = order
	}
	return repo
}

func (r *fakeOrderRepo) Checkout(context.Context, uint, string) (*model.Order, error) {
	return nil, nil
}

func (r *fakeOrderRepo) ListByUserID(context.Context, uint, int, int) ([]model.Order, int64, error) {
	return nil, 0, nil
}

func (r *fakeOrderRepo) ListAll(context.Context, int, int) ([]model.Order, int64, error) {
	return nil, 0, nil
}

func (r *fakeOrderRepo) FindByID(_ context.Context, id uint) (*model.Order, error) {
	return r.orders[id], nil
}

func (r *fakeOrderRepo) UpdateStatus(_ context.Context, id uint, status string) (*model.Order, error) {
	order := r.orders[id]
	order.Status = status
	return order, nil
}

func TestOrderService_UpdateStatus_AllowsValidForwardTransitions(t *testing.T) {
	cases := []struct {
		from, to string
	}{
		{model.OrderStatusPending, model.OrderStatusPaid},
		{model.OrderStatusPending, model.OrderStatusCancelled},
		{model.OrderStatusPaid, model.OrderStatusProcessing},
		{model.OrderStatusProcessing, model.OrderStatusShipped},
		{model.OrderStatusShipped, model.OrderStatusDelivered},
	}

	for _, tc := range cases {
		orderRepo := newFakeOrderRepo(&model.Order{ID: 1, Status: tc.from})
		audit := &fakeAuditLogRepo{}
		svc := NewOrderService(orderRepo, audit)

		updated, err := svc.UpdateStatus(context.Background(), 42, 1, tc.to)
		if err != nil {
			t.Errorf("%s -> %s: unexpected error: %v", tc.from, tc.to, err)
			continue
		}
		if updated.Status != tc.to {
			t.Errorf("%s -> %s: Status = %q, want %q", tc.from, tc.to, updated.Status, tc.to)
		}
		if len(audit.entries) != 1 || audit.entries[0].Action != "order.status_change" {
			t.Errorf("%s -> %s: expected 1 order.status_change audit entry, got %+v", tc.from, tc.to, audit.entries)
		}
	}
}

func TestOrderService_UpdateStatus_RejectsInvalidTransitions(t *testing.T) {
	cases := []struct {
		from, to string
	}{
		{model.OrderStatusDelivered, model.OrderStatusPending}, // terminal state
		{model.OrderStatusCancelled, model.OrderStatusPaid},    // terminal state
		{model.OrderStatusShipped, model.OrderStatusCancelled}, // can't cancel once shipped
		{model.OrderStatusPending, model.OrderStatusShipped},   // can't skip stages
		{model.OrderStatusPending, model.OrderStatusDelivered}, // can't skip stages
	}

	for _, tc := range cases {
		orderRepo := newFakeOrderRepo(&model.Order{ID: 1, Status: tc.from})
		audit := &fakeAuditLogRepo{}
		svc := NewOrderService(orderRepo, audit)

		_, err := svc.UpdateStatus(context.Background(), 42, 1, tc.to)
		if err == nil {
			t.Errorf("%s -> %s: expected error, got nil", tc.from, tc.to)
		}
		if len(audit.entries) != 0 {
			t.Errorf("%s -> %s: rejected transition must not write an audit entry, got %+v", tc.from, tc.to, audit.entries)
		}
	}
}

func TestOrderService_UpdateStatus_NotFound(t *testing.T) {
	orderRepo := newFakeOrderRepo()
	svc := NewOrderService(orderRepo, &fakeAuditLogRepo{})

	_, err := svc.UpdateStatus(context.Background(), 42, 999, model.OrderStatusPaid)
	if err == nil {
		t.Fatal("expected error for nonexistent order, got nil")
	}
}
