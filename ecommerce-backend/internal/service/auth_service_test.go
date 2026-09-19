package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/testutil"
)

// mockUserRepository is testify/mock-based — UserRepository has 4 methods
// (>3), same convention as mockPromotionRepository/mockReviewRepository.
type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}

func (m *mockUserRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	args := m.Called(ctx, id)
	user, _ := args.Get(0).(*model.User)
	return user, args.Error(1)
}

func (m *mockUserRepository) UpdatePassword(ctx context.Context, userID uint, passwordHash string) error {
	return m.Called(ctx, userID, passwordHash).Error(0)
}

func (m *mockUserRepository) List(ctx context.Context, search string, page, limit int) ([]model.User, int64, error) {
	args := m.Called(ctx, search, page, limit)
	users, _ := args.Get(0).([]model.User)
	return users, args.Get(1).(int64), args.Error(2)
}

func (m *mockUserRepository) UpdateRoles(ctx context.Context, userID uint, roles []model.Role) error {
	return m.Called(ctx, userID, roles).Error(0)
}

type mockPasswordResetStore struct {
	mock.Mock
}

func (m *mockPasswordResetStore) GenerateToken(ctx context.Context, userID uint, ttl time.Duration) (string, error) {
	args := m.Called(ctx, userID, ttl)
	return args.String(0), args.Error(1)
}

func (m *mockPasswordResetStore) Consume(ctx context.Context, token string) (uint, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(uint), args.Error(1)
}

func TestAuthService_ForgotPassword(t *testing.T) {
	t.Parallel()

	t.Run("existing email generates a reset token", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mockUserRepository)
		userRepo.On("FindByEmail", mock.Anything, "user@example.com").
			Return(&model.User{ID: 7, Email: "user@example.com"}, nil).Once()
		resetStore := new(mockPasswordResetStore)
		resetStore.On("GenerateToken", mock.Anything, uint(7), passwordResetTTL).Return("sometoken", nil).Once()
		svc := NewAuthService(userRepo, nil, nil, nil, resetStore, 4)

		err := svc.ForgotPassword(context.Background(), "user@example.com")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		userRepo.AssertExpectations(t)
		resetStore.AssertExpectations(t)
	})

	t.Run("unknown email does not generate a token or leak that fact", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mockUserRepository)
		userRepo.On("FindByEmail", mock.Anything, "ghost@example.com").Return(nil, nil).Once()
		resetStore := new(mockPasswordResetStore)
		svc := NewAuthService(userRepo, nil, nil, nil, resetStore, 4)

		err := svc.ForgotPassword(context.Background(), "ghost@example.com")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		resetStore.AssertNotCalled(t, "GenerateToken", mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestAuthService_ResetPassword(t *testing.T) {
	t.Parallel()

	t.Run("valid token and strong password succeeds", func(t *testing.T) {
		t.Parallel()

		resetStore := new(mockPasswordResetStore)
		resetStore.On("Consume", mock.Anything, "validtoken").Return(uint(7), nil).Once()
		userRepo := new(mockUserRepository)
		userRepo.On("UpdatePassword", mock.Anything, uint(7), mock.AnythingOfType("string")).Return(nil).Once()
		svc := NewAuthService(userRepo, nil, nil, nil, resetStore, 4)

		err := svc.ResetPassword(context.Background(), "validtoken", "NewPassw0rd")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		userRepo.AssertExpectations(t)
	})

	t.Run("invalid or expired token is rejected without touching the user repo", func(t *testing.T) {
		t.Parallel()

		resetStore := new(mockPasswordResetStore)
		resetStore.On("Consume", mock.Anything, "badtoken").Return(uint(0), nil).Once()
		userRepo := new(mockUserRepository)
		svc := NewAuthService(userRepo, nil, nil, nil, resetStore, 4)

		err := svc.ResetPassword(context.Background(), "badtoken", "NewPassw0rd")

		testutil.AssertAppError(t, err, apperror.CodeUnauthorized, 401)
		userRepo.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("weak password is rejected before the token is even consumed", func(t *testing.T) {
		t.Parallel()

		resetStore := new(mockPasswordResetStore)
		svc := NewAuthService(nil, nil, nil, nil, resetStore, 4)

		err := svc.ResetPassword(context.Background(), "validtoken", "weak")

		testutil.AssertAppError(t, err, apperror.CodeValidation, 400)
		resetStore.AssertNotCalled(t, "Consume", mock.Anything, mock.Anything)
	})
}
