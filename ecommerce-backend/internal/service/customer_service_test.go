package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/testutil"
)

// mockRoleRepository is testify/mock-based — RoleRepository now has 3
// methods (FindByName, List, FindByNames). Hand-rolled would be ≤3 by the
// old rule, but since we need behaviour-per-call control in UpdateUserRoles
// tests (different FindByNames return per scenario), testify/mock is clearer.
type mockRoleRepository struct {
	mock.Mock
}

func (m *mockRoleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	args := m.Called(ctx, name)
	role, _ := args.Get(0).(*model.Role)
	return role, args.Error(1)
}

func (m *mockRoleRepository) List(ctx context.Context) ([]model.Role, error) {
	args := m.Called(ctx)
	roles, _ := args.Get(0).([]model.Role)
	return roles, args.Error(1)
}

func (m *mockRoleRepository) FindByNames(ctx context.Context, names []string) ([]model.Role, error) {
	args := m.Called(ctx, names)
	roles, _ := args.Get(0).([]model.Role)
	return roles, args.Error(1)
}

func TestCustomerService_List(t *testing.T) {
	t.Parallel()

	userRepo := new(mockUserRepository)
	userRepo.On("List", mock.Anything, "budi", 1, 10).
		Return([]model.User{{ID: 1, Name: "Budi", Email: "budi@example.com"}}, int64(1), nil).Once()
	svc := NewCustomerService(userRepo, nil, nil)

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
		svc := NewCustomerService(userRepo, nil, nil)

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
		svc := NewCustomerService(userRepo, nil, nil)

		_, err := svc.GetByID(context.Background(), 999)

		testutil.AssertAppError(t, err, apperror.CodeNotFound, 404)
	})
}

func TestCustomerService_ListRoles(t *testing.T) {
	t.Parallel()

	roleRepo := new(mockRoleRepository)
	roleRepo.On("List", mock.Anything).Return([]model.Role{
		{ID: 1, Name: "admin"},
		{ID: 2, Name: "customer"},
	}, nil).Once()
	svc := NewCustomerService(nil, roleRepo, nil)

	roles, err := svc.ListRoles(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("len(roles) = %d, want 2", len(roles))
	}
	roleRepo.AssertExpectations(t)
}

func TestCustomerService_UpdateUserRoles(t *testing.T) {
	t.Parallel()

	adminRole := model.Role{ID: 1, Name: "admin"}
	customerRole := model.Role{ID: 2, Name: "customer"}

	t.Run("success — assign two roles and write audit log", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mockUserRepository)
		userRepo.On("FindByID", mock.Anything, uint(5)).Return(
			&model.User{ID: 5, Name: "Budi", Roles: []model.Role{customerRole}},
			nil,
		).Once()
		userRepo.On("UpdateRoles", mock.Anything, uint(5), []model.Role{adminRole, customerRole}).Return(nil).Once()

		roleRepo := new(mockRoleRepository)
		roleRepo.On("FindByNames", mock.Anything, []string{"admin", "customer"}).
			Return([]model.Role{adminRole, customerRole}, nil).Once()

		audit := &fakeAuditLogRepo{}
		svc := NewCustomerService(userRepo, roleRepo, audit)

		err := svc.UpdateUserRoles(context.Background(), 99, 5, []string{"admin", "customer"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(audit.entries) != 1 {
			t.Fatalf("audit entries = %d, want 1", len(audit.entries))
		}
		if audit.entries[0].Action != "user.roles_changed" {
			t.Errorf("audit action = %q, want \"user.roles_changed\"", audit.entries[0].Action)
		}
		userRepo.AssertExpectations(t)
		roleRepo.AssertExpectations(t)
	})

	t.Run("unknown role name returns validation error", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mockUserRepository)
		userRepo.On("FindByID", mock.Anything, uint(5)).Return(
			&model.User{ID: 5, Roles: []model.Role{customerRole}},
			nil,
		).Once()

		roleRepo := new(mockRoleRepository)
		// Only 1 role found for 2 submitted names → mismatch triggers validation error.
		roleRepo.On("FindByNames", mock.Anything, []string{"admin", "nonexistent"}).
			Return([]model.Role{adminRole}, nil).Once()

		svc := NewCustomerService(userRepo, roleRepo, &fakeAuditLogRepo{})

		err := svc.UpdateUserRoles(context.Background(), 99, 5, []string{"admin", "nonexistent"})

		testutil.AssertAppError(t, err, apperror.CodeValidation, 400)
	})

	t.Run("target user not found returns 404", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mockUserRepository)
		userRepo.On("FindByID", mock.Anything, uint(404)).Return(nil, nil).Once()

		svc := NewCustomerService(userRepo, new(mockRoleRepository), &fakeAuditLogRepo{})

		err := svc.UpdateUserRoles(context.Background(), 99, 404, []string{"customer"})

		testutil.AssertAppError(t, err, apperror.CodeNotFound, 404)
	})

	t.Run("admin removing own admin role returns conflict", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mockUserRepository)
		userRepo.On("FindByID", mock.Anything, uint(1)).Return(
			&model.User{ID: 1, Roles: []model.Role{adminRole}},
			nil,
		).Once()

		roleRepo := new(mockRoleRepository)
		roleRepo.On("FindByNames", mock.Anything, []string{"customer"}).
			Return([]model.Role{customerRole}, nil).Once()

		svc := NewCustomerService(userRepo, roleRepo, &fakeAuditLogRepo{})

		// callerID == targetUserID == 1, and new roles do not include "admin".
		err := svc.UpdateUserRoles(context.Background(), 1, 1, []string{"customer"})

		testutil.AssertAppError(t, err, apperror.CodeConflict, 409)
		userRepo.AssertExpectations(t)
	})

	t.Run("admin updating another user without admin role is allowed", func(t *testing.T) {
		t.Parallel()

		userRepo := new(mockUserRepository)
		// callerID (1) != targetUserID (5) — self-lockout guard does not apply.
		userRepo.On("FindByID", mock.Anything, uint(5)).Return(
			&model.User{ID: 5, Roles: []model.Role{adminRole, customerRole}},
			nil,
		).Once()
		userRepo.On("UpdateRoles", mock.Anything, uint(5), []model.Role{customerRole}).Return(nil).Once()

		roleRepo := new(mockRoleRepository)
		roleRepo.On("FindByNames", mock.Anything, []string{"customer"}).
			Return([]model.Role{customerRole}, nil).Once()

		svc := NewCustomerService(userRepo, roleRepo, &fakeAuditLogRepo{})

		err := svc.UpdateUserRoles(context.Background(), 1, 5, []string{"customer"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		userRepo.AssertExpectations(t)
	})
}
