package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/testutil"
)

func TestCustomerService_List(t *testing.T) {
	t.Parallel()

	userRepo := new(mockUserRepository)
	userRepo.On("List", mock.Anything, "budi", 1, 10).
		Return([]model.User{{ID: 1, Name: "Budi", Email: "budi@example.com"}}, int64(1), nil).Once()
	svc := NewCustomerService(userRepo)

	customers, meta, err := svc.List(context.Background(), model.CustomerListQuery{Search: "budi"})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(customers) != 1 || customers[0].Name != "Budi" {
		t.Errorf("customers = %+v, want one customer named Budi", customers)
	}
	if meta.Total != 1 {
		t.Errorf("meta.Total = %d, want 1", meta.Total)
	}
	userRepo.AssertExpectations(t)
}

func TestCustomerService_GetByID(t *testing.T) {
	t.Parallel()

	t.Run("found", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mockUserRepository)
		userRepo.On("FindByID", mock.Anything, uint(1)).Return(&model.User{ID: 1, Name: "Budi"}, nil).Once()
		svc := NewCustomerService(userRepo)

		customer, err := svc.GetByID(context.Background(), 1)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if customer.ID != 1 {
			t.Errorf("ID = %d, want 1", customer.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mockUserRepository)
		userRepo.On("FindByID", mock.Anything, uint(999)).Return(nil, nil).Once()
		svc := NewCustomerService(userRepo)

		_, err := svc.GetByID(context.Background(), 999)

		testutil.AssertAppError(t, err, apperror.CodeNotFound, 404)
	})
}
