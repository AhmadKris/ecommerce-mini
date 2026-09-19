package service

import (
	"context"
	"testing"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/testutil"
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

func (r *fakeOrderRepo) Checkout(context.Context, uint, string, string) (*model.Order, error) {
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

func TestOrderService_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		orderRepo := newFakeOrderRepo(&model.Order{ID: 1, Status: model.OrderStatusPaid})
		svc := NewOrderService(orderRepo, &fakeAuditLogRepo{})

		order, err := svc.GetByID(context.Background(), 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if order.ID != 1 {
			t.Errorf("ID = %d, want 1", order.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		orderRepo := newFakeOrderRepo()
		svc := NewOrderService(orderRepo, &fakeAuditLogRepo{})

		_, err := svc.GetByID(context.Background(), 999)
		testutil.AssertAppError(t, err, apperror.CodeNotFound, 404)
	})
}

func TestOrderService_UpdateStatus_AllowsValidForwardTransitions(t *testing.T) {
	t.Parallel()

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
		t.Run(tc.from+"->"+tc.to, func(t *testing.T) {
			t.Parallel()

			orderRepo := newFakeOrderRepo(&model.Order{ID: 1, Status: tc.from})
			audit := &fakeAuditLogRepo{}
			svc := NewOrderService(orderRepo, audit)

			updated, err := svc.UpdateStatus(context.Background(), 42, 1, tc.to)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if updated.Status != tc.to {
				t.Errorf("Status = %q, want %q", updated.Status, tc.to)
			}
			if len(audit.entries) != 1 || audit.entries[0].Action != "order.status_change" {
				t.Errorf("expected 1 order.status_change audit entry, got %+v", audit.entries)
			}
		})
	}
}

func TestOrderService_UpdateStatus_RejectsInvalidTransitions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		from, to string
	}{
		{"terminal state delivered", model.OrderStatusDelivered, model.OrderStatusPending},
		{"terminal state cancelled", model.OrderStatusCancelled, model.OrderStatusPaid},
		{"cannot cancel once shipped", model.OrderStatusShipped, model.OrderStatusCancelled},
		{"cannot skip stages (pending->shipped)", model.OrderStatusPending, model.OrderStatusShipped},
		{"cannot skip stages (pending->delivered)", model.OrderStatusPending, model.OrderStatusDelivered},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			orderRepo := newFakeOrderRepo(&model.Order{ID: 1, Status: tc.from})
			audit := &fakeAuditLogRepo{}
			svc := NewOrderService(orderRepo, audit)

			_, err := svc.UpdateStatus(context.Background(), 42, 1, tc.to)
			testutil.AssertAppError(t, err, apperror.CodeConflict, 409)
			if len(audit.entries) != 0 {
				t.Errorf("rejected transition must not write an audit entry, got %+v", audit.entries)
			}
		})
	}
}

func TestOrderService_UpdateStatus_NotFound(t *testing.T) {
	t.Parallel()

	orderRepo := newFakeOrderRepo()
	svc := NewOrderService(orderRepo, &fakeAuditLogRepo{})

	_, err := svc.UpdateStatus(context.Background(), 42, 999, model.OrderStatusPaid)
	testutil.AssertAppError(t, err, apperror.CodeNotFound, 404)
}
