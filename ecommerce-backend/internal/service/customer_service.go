package service

import (
	"context"
	"fmt"
	"log/slog"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// CustomerService implements admin read access to registered accounts and
// role management. List/detail are read-only; UpdateUserRoles is the single
// mutation endpoint — it replaces the user's full role set.
type CustomerService struct {
	userRepo     repository.UserRepository
	roleRepo     repository.RoleRepository
	auditLogRepo repository.AuditLogRepository
}

// NewCustomerService builds a CustomerService with its dependencies.
func NewCustomerService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	auditLogRepo repository.AuditLogRepository,
) *CustomerService {
	return &CustomerService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		auditLogRepo: auditLogRepo,
	}
}

// List returns a page of registered accounts, optionally filtered by a
// name/email search, newest first.
func (s *CustomerService) List(ctx context.Context, query model.CustomerListQuery) ([]model.User, model.Meta, error) {
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

	users, total, err := s.userRepo.List(ctx, query.Search, page, limit)
	if err != nil {
		return nil, model.Meta{}, apperror.Internal(fmt.Errorf("service: list customers: %w", err))
	}
	return users, model.NewMeta(page, limit, total), nil
}

// GetByID returns a single account's profile and roles.
func (s *CustomerService) GetByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: get customer: %w", err))
	}
	if user == nil {
		return nil, apperror.NotFound("Customer tidak ditemukan", nil)
	}
	return user, nil
}

// ListRoles returns all available roles with their permissions. Used by the
// admin UI to populate the role selection checkboxes.
func (s *CustomerService) ListRoles(ctx context.Context) ([]model.Role, error) {
	roles, err := s.roleRepo.List(ctx)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: list roles: %w", err))
	}
	return roles, nil
}

// UpdateUserRoles replaces the full role set for targetUserID. callerID is
// needed for the self-lockout guard: an admin cannot remove their own admin
// role, because that would immediately revoke access to the management UI.
//
// The audit log is written best-effort after the write succeeds — consistent
// with ProductService/CategoryService. A failed audit write is logged but does
// not roll back the already-persisted role change.
func (s *CustomerService) UpdateUserRoles(ctx context.Context, callerID, targetUserID uint, roleNames []string) error {
	target, err := s.userRepo.FindByID(ctx, targetUserID)
	if err != nil {
		return apperror.Internal(fmt.Errorf("service: update user roles: %w", err))
	}
	if target == nil {
		return apperror.NotFound("Customer tidak ditemukan", nil)
	}

	// Validate that every submitted name maps to a real role.
	roles, err := s.roleRepo.FindByNames(ctx, roleNames)
	if err != nil {
		return apperror.Internal(fmt.Errorf("service: update user roles: %w", err))
	}
	if len(roles) != len(roleNames) {
		return apperror.Validation("Satu atau lebih nama role tidak valid", []string{
			"role_names: contains unknown role name(s)",
		})
	}

	// Self-lockout guard: prevent an admin from revoking their own admin role.
	// The check is done in the service layer (not the handler) so it's always
	// enforced regardless of caller, and matches the pattern used for
	// ownership checks throughout the codebase.
	if callerID == targetUserID && !containsName(roleNames, "admin") {
		return apperror.Conflict(
			"Admin tidak dapat menghapus role admin dari diri sendiri",
			fmt.Errorf("service: update user roles: self-lockout prevented"),
		)
	}

	beforeRoles := target.RoleNames()

	if err := s.userRepo.UpdateRoles(ctx, targetUserID, roles); err != nil {
		return apperror.Internal(fmt.Errorf("service: update user roles: %w", err))
	}

	afterRoles := make([]string, len(roles))
	for i, r := range roles {
		afterRoles[i] = r.Name
	}

	// Best-effort audit log — failure here does not undo the role change.
	if err := s.auditLogRepo.Create(ctx, &model.AuditLog{
		ActorID:    callerID,
		Action:     "user.roles_changed",
		Resource:   "user",
		ResourceID: targetUserID,
		Metadata:   fmt.Sprintf(`{"before":%q,"after":%q}`, beforeRoles, afterRoles),
	}); err != nil {
		slog.Error("failed to write audit log for user.roles_changed",
			slog.Uint64("actor_id", uint64(callerID)),
			slog.Uint64("target_user_id", uint64(targetUserID)),
			slog.Any("err", err),
		)
	}

	return nil
}

// containsName reports whether name appears in the slice.
func containsName(names []string, name string) bool {
	for _, n := range names {
		if n == name {
			return true
		}
	}
	return false
}
