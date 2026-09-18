package service

import (
	"context"
	"testing"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

type fakeInventoryRepo struct {
	stock       map[uint]int
	nextMoveID  uint
	adjustCalls []model.AdjustInventoryRequest
}

func newFakeInventoryRepo(stock map[uint]int) *fakeInventoryRepo {
	return &fakeInventoryRepo{stock: stock, nextMoveID: 1}
}

func (r *fakeInventoryRepo) Adjust(_ context.Context, actorID uint, req model.AdjustInventoryRequest) (*model.InventoryMovement, error) {
	r.adjustCalls = append(r.adjustCalls, req)

	before, ok := r.stock[req.ProductID]
	if !ok {
		return nil, repository.ErrProductNotFound
	}

	after := before
	switch req.Type {
	case model.InventoryMovementTypeIn:
		after = before + req.Quantity
	case model.InventoryMovementTypeOut:
		after = before - req.Quantity
	case model.InventoryMovementTypeCorrection:
		after = req.Quantity
	}
	if after < 0 {
		return nil, repository.ErrInsufficientStock
	}

	r.stock[req.ProductID] = after
	movement := &model.InventoryMovement{
		ID: r.nextMoveID, ProductID: req.ProductID, Type: req.Type,
		Quantity: req.Quantity, BeforeQuantity: before, AfterQuantity: after,
		Reason: req.Reason, Reference: req.Reference, CreatedBy: actorID,
	}
	r.nextMoveID++
	return movement, nil
}

func (r *fakeInventoryRepo) ListByProductID(context.Context, uint, int, int) ([]model.InventoryMovement, int64, error) {
	return nil, 0, nil
}

func TestInventoryService_Adjust_AppliesInOutAndCorrection(t *testing.T) {
	repo := newFakeInventoryRepo(map[uint]int{1: 10})
	svc := NewInventoryService(repo)

	movement, err := svc.Adjust(context.Background(), 42, model.AdjustInventoryRequest{
		ProductID: 1, Type: model.InventoryMovementTypeIn, Quantity: 5, Reason: "restock",
	})
	if err != nil {
		t.Fatalf("in: unexpected error: %v", err)
	}
	if movement.AfterQuantity != 15 {
		t.Errorf("in: AfterQuantity = %d, want 15", movement.AfterQuantity)
	}

	movement, err = svc.Adjust(context.Background(), 42, model.AdjustInventoryRequest{
		ProductID: 1, Type: model.InventoryMovementTypeOut, Quantity: 3, Reason: "damaged",
	})
	if err != nil {
		t.Fatalf("out: unexpected error: %v", err)
	}
	if movement.AfterQuantity != 12 {
		t.Errorf("out: AfterQuantity = %d, want 12", movement.AfterQuantity)
	}

	movement, err = svc.Adjust(context.Background(), 42, model.AdjustInventoryRequest{
		ProductID: 1, Type: model.InventoryMovementTypeCorrection, Quantity: 100, Reason: "stock opname",
	})
	if err != nil {
		t.Fatalf("correction: unexpected error: %v", err)
	}
	if movement.AfterQuantity != 100 {
		t.Errorf("correction: AfterQuantity = %d, want 100 (absolute, not delta)", movement.AfterQuantity)
	}
}

func TestInventoryService_Adjust_RejectsOutBelowZero(t *testing.T) {
	repo := newFakeInventoryRepo(map[uint]int{1: 5})
	svc := NewInventoryService(repo)

	_, err := svc.Adjust(context.Background(), 42, model.AdjustInventoryRequest{
		ProductID: 1, Type: model.InventoryMovementTypeOut, Quantity: 10, Reason: "damaged",
	})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeConflict {
		t.Fatalf("error = %v, want apperror Conflict", err)
	}
}

func TestInventoryService_Adjust_ProductNotFound(t *testing.T) {
	repo := newFakeInventoryRepo(map[uint]int{})
	svc := NewInventoryService(repo)

	_, err := svc.Adjust(context.Background(), 42, model.AdjustInventoryRequest{
		ProductID: 999, Type: model.InventoryMovementTypeIn, Quantity: 1, Reason: "restock",
	})
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeNotFound {
		t.Fatalf("error = %v, want apperror NotFound", err)
	}
}
