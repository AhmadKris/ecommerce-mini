package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/testutil"
)

// mockPromotionRepository is testify/mock-based — PromotionRepository has 5
// methods (>3), same convention as mockCategoryRepository.
type mockPromotionRepository struct {
	mock.Mock
}

func (m *mockPromotionRepository) Create(ctx context.Context, promotion *model.Promotion) error {
	return m.Called(ctx, promotion).Error(0)
}

func (m *mockPromotionRepository) Update(ctx context.Context, promotion *model.Promotion) error {
	return m.Called(ctx, promotion).Error(0)
}

func (m *mockPromotionRepository) Delete(ctx context.Context, id uint) error {
	return m.Called(ctx, id).Error(0)
}

func (m *mockPromotionRepository) FindByID(ctx context.Context, id uint) (*model.Promotion, error) {
	args := m.Called(ctx, id)
	promotion, _ := args.Get(0).(*model.Promotion)
	return promotion, args.Error(1)
}

func (m *mockPromotionRepository) List(ctx context.Context) ([]model.Promotion, error) {
	args := m.Called(ctx)
	promotions, _ := args.Get(0).([]model.Promotion)
	return promotions, args.Error(1)
}

func validCreatePromotionRequest() model.CreatePromotionRequest {
	return model.CreatePromotionRequest{
		Code: "save10", Type: model.PromotionTypePercentage, Value: 10,
		UsageLimit: 100, StartsAt: time.Now(), EndsAt: time.Now().Add(24 * time.Hour),
	}
}

func TestPromotionService_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		setupMock  func(*mockPromotionRepository)
		wantErr    bool
		wantCode   apperror.Code
		wantStatus int
	}{
		{
			name: "success normalizes code to uppercase and records audit",
			setupMock: func(m *mockPromotionRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(p *model.Promotion) bool {
					return p.Code == "SAVE10"
				})).Return(nil).Once()
			},
		},
		{
			name: "duplicate code",
			setupMock: func(m *mockPromotionRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return(repository.ErrPromotionCodeTaken).Once()
			},
			wantErr: true, wantCode: apperror.CodeDuplicateEntry, wantStatus: 409,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			promotionRepo := new(mockPromotionRepository)
			tc.setupMock(promotionRepo)
			audit := &fakeAuditLogRepo{}
			svc := NewPromotionService(promotionRepo, audit)

			_, err := svc.Create(context.Background(), 42, validCreatePromotionRequest())

			if tc.wantErr {
				testutil.AssertAppError(t, err, tc.wantCode, tc.wantStatus)
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(audit.entries) != 1 || audit.entries[0].Action != "promotion.create" {
					t.Errorf("expected 1 promotion.create audit entry, got %+v", audit.entries)
				}
			}
			promotionRepo.AssertExpectations(t)
		})
	}
}

func TestPromotionService_Update_NotFound(t *testing.T) {
	t.Parallel()

	promotionRepo := new(mockPromotionRepository)
	promotionRepo.On("FindByID", mock.Anything, uint(1)).Return(nil, nil).Once()
	svc := NewPromotionService(promotionRepo, &fakeAuditLogRepo{})

	req := model.UpdatePromotionRequest{
		Type: model.PromotionTypeFixed, Value: 5000, UsageLimit: 10,
		StartsAt: time.Now(), EndsAt: time.Now().Add(time.Hour), Status: model.PromotionStatusActive,
	}
	_, err := svc.Update(context.Background(), 42, 1, req)

	testutil.AssertAppError(t, err, apperror.CodeNotFound, 404)
	promotionRepo.AssertExpectations(t)
}

func TestPromotionService_Delete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		repoErr    error
		wantErr    bool
		wantCode   apperror.Code
		wantStatus int
	}{
		{name: "success"},
		{name: "not found", repoErr: gorm.ErrRecordNotFound, wantErr: true, wantCode: apperror.CodeNotFound, wantStatus: 404},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			promotionRepo := new(mockPromotionRepository)
			promotionRepo.On("Delete", mock.Anything, uint(1)).Return(tc.repoErr).Once()
			svc := NewPromotionService(promotionRepo, &fakeAuditLogRepo{})

			err := svc.Delete(context.Background(), 42, 1)

			if tc.wantErr {
				testutil.AssertAppError(t, err, tc.wantCode, tc.wantStatus)
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			promotionRepo.AssertExpectations(t)
		})
	}
}
