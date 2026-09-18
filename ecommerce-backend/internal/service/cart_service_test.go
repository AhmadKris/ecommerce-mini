package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/testutil"
)

// mockCartRepository is testify/mock-based — CartRepository has 7 methods
// (>3), same convention as mockCategoryRepository.
type mockCartRepository struct {
	mock.Mock
}

func (m *mockCartRepository) FindOrCreateByUserID(ctx context.Context, userID uint) (*model.Cart, error) {
	args := m.Called(ctx, userID)
	cart, _ := args.Get(0).(*model.Cart)
	return cart, args.Error(1)
}

func (m *mockCartRepository) FindItemByID(ctx context.Context, itemID uint) (*model.CartItem, error) {
	args := m.Called(ctx, itemID)
	item, _ := args.Get(0).(*model.CartItem)
	return item, args.Error(1)
}

func (m *mockCartRepository) FindItemByProductID(ctx context.Context, cartID, productID uint) (*model.CartItem, error) {
	args := m.Called(ctx, cartID, productID)
	item, _ := args.Get(0).(*model.CartItem)
	return item, args.Error(1)
}

func (m *mockCartRepository) CreateItem(ctx context.Context, item *model.CartItem) error {
	return m.Called(ctx, item).Error(0)
}

func (m *mockCartRepository) UpdateItemQuantity(ctx context.Context, item *model.CartItem) error {
	return m.Called(ctx, item).Error(0)
}

func (m *mockCartRepository) DeleteItem(ctx context.Context, itemID uint) error {
	return m.Called(ctx, itemID).Error(0)
}

func (m *mockCartRepository) ListItems(ctx context.Context, cartID uint) ([]model.CartItem, error) {
	args := m.Called(ctx, cartID)
	items, _ := args.Get(0).([]model.CartItem)
	return items, args.Error(1)
}

var testCart = &model.Cart{ID: 1, UserID: 42}

func TestCartService_AddItem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		quantity   int
		setupMock  func(*mockCartRepository)
		wantErr    bool
		wantCode   apperror.Code
		wantStatus int
	}{
		{
			name:     "new item within stock succeeds",
			quantity: 3,
			setupMock: func(m *mockCartRepository) {
				m.On("FindOrCreateByUserID", mock.Anything, uint(42)).Return(testCart, nil)
				m.On("FindItemByProductID", mock.Anything, testCart.ID, uint(1)).Return(nil, nil).Once()
				m.On("CreateItem", mock.Anything, mock.Anything).Return(nil).Once()
				m.On("ListItems", mock.Anything, testCart.ID).Return([]model.CartItem{}, nil)
			},
		},
		{
			name:     "existing item increments quantity",
			quantity: 2,
			setupMock: func(m *mockCartRepository) {
				existing := testutil.NewCartItem(testutil.WithCartItemQuantity(1))
				m.On("FindOrCreateByUserID", mock.Anything, uint(42)).Return(testCart, nil)
				m.On("FindItemByProductID", mock.Anything, testCart.ID, uint(1)).Return(existing, nil).Once()
				m.On("UpdateItemQuantity", mock.Anything, mock.MatchedBy(func(i *model.CartItem) bool {
					return i.Quantity == 3 // 1 existing + 2 requested
				})).Return(nil).Once()
				m.On("ListItems", mock.Anything, testCart.ID).Return([]model.CartItem{}, nil)
			},
		},
		{
			name:     "requested quantity exceeds stock",
			quantity: 999,
			setupMock: func(m *mockCartRepository) {
				m.On("FindOrCreateByUserID", mock.Anything, uint(42)).Return(testCart, nil)
				m.On("FindItemByProductID", mock.Anything, testCart.ID, uint(1)).Return(nil, nil).Once()
			},
			wantErr: true, wantCode: apperror.CodeConflict, wantStatus: 409,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cartRepo := new(mockCartRepository)
			tc.setupMock(cartRepo)
			productRepo := newFakeProductRepo()
			product := testutil.NewProduct(testutil.WithStock(10))
			productRepo.products[product.ID] = product
			svc := NewCartService(cartRepo, productRepo)

			_, err := svc.AddItem(context.Background(), 42, model.AddCartItemRequest{ProductID: 1, Quantity: tc.quantity})

			if tc.wantErr {
				testutil.AssertAppError(t, err, tc.wantCode, tc.wantStatus)
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			cartRepo.AssertExpectations(t)
		})
	}
}

func TestCartService_UpdateItemQuantity_OwnershipAndStock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMock  func(*mockCartRepository)
		wantErr    bool
		wantCode   apperror.Code
		wantStatus int
	}{
		{
			name: "owner updating within stock succeeds",
			setupMock: func(m *mockCartRepository) {
				item := testutil.NewCartItem(testutil.WithCartItemID(5))
				m.On("FindOrCreateByUserID", mock.Anything, uint(42)).Return(testCart, nil)
				m.On("FindItemByID", mock.Anything, uint(5)).Return(item, nil)
				m.On("UpdateItemQuantity", mock.Anything, mock.Anything).Return(nil).Once()
				m.On("ListItems", mock.Anything, testCart.ID).Return([]model.CartItem{}, nil)
			},
		},
		{
			name: "item belongs to a different cart reports NotFound, not Forbidden",
			setupMock: func(m *mockCartRepository) {
				item := testutil.NewCartItem(testutil.WithCartItemID(5))
				item.CartID = 999 // someone else's cart
				m.On("FindOrCreateByUserID", mock.Anything, uint(42)).Return(testCart, nil)
				m.On("FindItemByID", mock.Anything, uint(5)).Return(item, nil)
			},
			wantErr: true, wantCode: apperror.CodeNotFound, wantStatus: 404,
		},
		{
			name: "item does not exist at all",
			setupMock: func(m *mockCartRepository) {
				m.On("FindOrCreateByUserID", mock.Anything, uint(42)).Return(testCart, nil)
				m.On("FindItemByID", mock.Anything, uint(5)).Return(nil, nil)
			},
			wantErr: true, wantCode: apperror.CodeNotFound, wantStatus: 404,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			cartRepo := new(mockCartRepository)
			tc.setupMock(cartRepo)
			productRepo := newFakeProductRepo()
			svc := NewCartService(cartRepo, productRepo)

			_, err := svc.UpdateItemQuantity(context.Background(), 42, 5, model.UpdateCartItemRequest{Quantity: 2})

			if tc.wantErr {
				testutil.AssertAppError(t, err, tc.wantCode, tc.wantStatus)
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			cartRepo.AssertExpectations(t)
		})
	}
}

// BenchmarkBuildCartResponse covers the hot path called on every cart
// read/mutation (Get, AddItem, UpdateItemQuantity, RemoveItem all end with
// it) — run against an in-memory slice, not a database, so the number
// reflects this function's own overhead only.
func BenchmarkBuildCartResponse(b *testing.B) {
	items := make([]model.CartItem, 20)
	for i := range items {
		product := testutil.NewProduct(testutil.WithProductID(uint(i + 1)))
		items[i] = *testutil.NewCartItem(testutil.WithCartItemID(uint(i+1)), testutil.WithCartItemProduct(product))
	}

	b.ResetTimer()
	for b.Loop() {
		buildCartResponse(items)
	}
}
