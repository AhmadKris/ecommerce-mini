package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/testutil"
)

// mockReviewRepository is testify/mock-based — ReviewRepository has 6
// methods (>3), same convention as mockPromotionRepository.
type mockReviewRepository struct {
	mock.Mock
}

func (m *mockReviewRepository) Create(ctx context.Context, review *model.Review) error {
	return m.Called(ctx, review).Error(0)
}

func (m *mockReviewRepository) FindByID(ctx context.Context, id uint) (*model.Review, error) {
	args := m.Called(ctx, id)
	review, _ := args.Get(0).(*model.Review)
	return review, args.Error(1)
}

func (m *mockReviewRepository) UpdateStatus(ctx context.Context, id uint, status string) (*model.Review, error) {
	args := m.Called(ctx, id, status)
	review, _ := args.Get(0).(*model.Review)
	return review, args.Error(1)
}

func (m *mockReviewRepository) ListByProductID(ctx context.Context, productID uint, status string, page, limit int) ([]model.Review, int64, error) {
	args := m.Called(ctx, productID, status, page, limit)
	reviews, _ := args.Get(0).([]model.Review)
	return reviews, args.Get(1).(int64), args.Error(2)
}

func (m *mockReviewRepository) ListAll(ctx context.Context, page, limit int) ([]model.Review, int64, error) {
	args := m.Called(ctx, page, limit)
	reviews, _ := args.Get(0).([]model.Review)
	return reviews, args.Get(1).(int64), args.Error(2)
}

func (m *mockReviewRepository) FindDeliveredOrderID(ctx context.Context, userID, productID uint) (uint, error) {
	args := m.Called(ctx, userID, productID)
	return args.Get(0).(uint), args.Error(1)
}

func validCreateReviewRequest() model.CreateReviewRequest {
	return model.CreateReviewRequest{Rating: 5, Title: "Barang bagus", Body: "Sangat puas dengan kualitas produknya."}
}

func TestReviewService_Create(t *testing.T) {
	t.Parallel()

	product := &model.Product{ID: 7, Slug: "kopi-hitam"}

	tests := []struct {
		name       string
		setupMock  func(*mockReviewRepository)
		wantErr    bool
		wantCode   apperror.Code
		wantStatus int
	}{
		{
			name: "success when caller has a delivered order for the product",
			setupMock: func(m *mockReviewRepository) {
				m.On("FindDeliveredOrderID", mock.Anything, uint(42), uint(7)).Return(uint(99), nil).Once()
				m.On("Create", mock.Anything, mock.MatchedBy(func(r *model.Review) bool {
					return r.ProductID == 7 && r.UserID == 42 && r.OrderID == 99 && r.Status == model.ReviewStatusPending
				})).Return(nil).Once()
			},
		},
		{
			name: "rejected when caller never received the product",
			setupMock: func(m *mockReviewRepository) {
				m.On("FindDeliveredOrderID", mock.Anything, uint(42), uint(7)).Return(uint(0), nil).Once()
			},
			wantErr: true, wantCode: apperror.CodeForbidden, wantStatus: 403,
		},
		{
			name: "rejected on duplicate review",
			setupMock: func(m *mockReviewRepository) {
				m.On("FindDeliveredOrderID", mock.Anything, uint(42), uint(7)).Return(uint(99), nil).Once()
				m.On("Create", mock.Anything, mock.Anything).Return(repository.ErrReviewAlreadyExists).Once()
			},
			wantErr: true, wantCode: apperror.CodeDuplicateEntry, wantStatus: 409,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			productRepo := newFakeProductRepo()
			productRepo.products[product.ID] = product

			reviewRepo := new(mockReviewRepository)
			tc.setupMock(reviewRepo)
			audit := &fakeAuditLogRepo{}
			svc := NewReviewService(reviewRepo, productRepo, audit)

			_, err := svc.Create(context.Background(), 42, "kopi-hitam", validCreateReviewRequest())

			if tc.wantErr {
				testutil.AssertAppError(t, err, tc.wantCode, tc.wantStatus)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(audit.entries) != 1 || audit.entries[0].Action != "review.create" {
					t.Errorf("expected 1 review.create audit entry, got %+v", audit.entries)
				}
			}
			reviewRepo.AssertExpectations(t)
		})
	}
}

func TestReviewService_Create_ProductNotFound(t *testing.T) {
	t.Parallel()

	productRepo := newFakeProductRepo()
	svc := NewReviewService(new(mockReviewRepository), productRepo, &fakeAuditLogRepo{})

	_, err := svc.Create(context.Background(), 42, "tidak-ada", validCreateReviewRequest())

	testutil.AssertAppError(t, err, apperror.CodeNotFound, 404)
}

func TestReviewService_UpdateStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		existing   *model.Review
		wantErr    bool
		wantCode   apperror.Code
		wantStatus int
	}{
		{
			name:     "approve moves a pending review",
			existing: &model.Review{ID: 1, Status: model.ReviewStatusPending},
		},
		{
			name:       "not found",
			existing:   nil,
			wantErr:    true,
			wantCode:   apperror.CodeNotFound,
			wantStatus: 404,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			reviewRepo := new(mockReviewRepository)
			reviewRepo.On("FindByID", mock.Anything, uint(1)).Return(tc.existing, nil).Once()
			if tc.existing != nil {
				updated := &model.Review{ID: 1, Status: model.ReviewStatusApproved}
				reviewRepo.On("UpdateStatus", mock.Anything, uint(1), model.ReviewStatusApproved).Return(updated, nil).Once()
			}
			audit := &fakeAuditLogRepo{}
			svc := NewReviewService(reviewRepo, newFakeProductRepo(), audit)

			_, err := svc.UpdateStatus(context.Background(), 1, 1, model.ReviewStatusApproved)

			if tc.wantErr {
				testutil.AssertAppError(t, err, tc.wantCode, tc.wantStatus)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(audit.entries) != 1 || audit.entries[0].Action != "review.status_change" {
					t.Errorf("expected 1 review.status_change audit entry, got %+v", audit.entries)
				}
			}
			reviewRepo.AssertExpectations(t)
		})
	}
}
