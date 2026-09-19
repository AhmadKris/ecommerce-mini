package service

import (
	"context"
	"fmt"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// CustomerService implements admin read access to registered accounts —
// list and detail only. There is no role/permission editing here yet (see
// .claude/CLAUDE.md Known Issues: "add it when a real need exists," and
// none has come up beyond viewing who's registered).
type CustomerService struct {
	userRepo repository.UserRepository
}

// NewCustomerService builds a CustomerService with its dependencies.
func NewCustomerService(userRepo repository.UserRepository) *CustomerService {
	return &CustomerService{userRepo: userRepo}
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
