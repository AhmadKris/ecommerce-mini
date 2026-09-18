// Package service holds business logic per resource: stock checks at
// checkout, email uniqueness at registration, multi-repository orchestration
// inside transactions. It translates repository errors into
// *apperror.AppError before they reach the handler layer.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"
	"unicode"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// ErrInvalidCredentials means the email/password pair did not match. It is
// never distinguished by field in the response, so a client can't use the
// error to enumerate registered emails.
var ErrInvalidCredentials = errors.New("invalid email or password")

// ErrRefreshTokenReused means a refresh token was presented after it had
// already been rotated (or otherwise blacklisted).
var ErrRefreshTokenReused = errors.New("refresh token already used or revoked")

const customerRoleName = "customer"

// AuthService implements registration, login, and refresh-token rotation.
type AuthService struct {
	userRepo   repository.UserRepository
	roleRepo   repository.RoleRepository
	tokens     *auth.TokenManager
	blacklist  *auth.RefreshBlacklist
	bcryptCost int
}

// NewAuthService builds an AuthService with its dependencies.
func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	tokens *auth.TokenManager,
	blacklist *auth.RefreshBlacklist,
	bcryptCost int,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		tokens:     tokens,
		blacklist:  blacklist,
		bcryptCost: bcryptCost,
	}
}

// Register creates a new customer account. Email uniqueness is enforced by
// the database constraint (see repository.ErrEmailTaken), not a pre-check,
// so two concurrent registrations for the same email can't both succeed.
func (s *AuthService) Register(ctx context.Context, req model.RegisterRequest) (*model.User, error) {
	if err := validatePasswordStrength(req.Password); err != nil {
		return nil, err
	}

	passwordHash, err := auth.HashPassword(req.Password, s.bcryptCost)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: register: %w", err))
	}

	customerRole, err := s.roleRepo.FindByName(ctx, customerRoleName)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: register: %w", err))
	}
	if customerRole == nil {
		return nil, apperror.Internal(fmt.Errorf("service: register: role %q not seeded", customerRoleName))
	}

	user := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Roles:        []model.Role{*customerRole},
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		if errors.Is(err, repository.ErrEmailTaken) {
			return nil, apperror.DuplicateEntry("Email sudah terdaftar", err)
		}
		return nil, apperror.Internal(fmt.Errorf("service: register: %w", err))
	}

	return user, nil
}

// Login verifies credentials and issues a fresh access/refresh token pair.
func (s *AuthService) Login(ctx context.Context, req model.LoginRequest) (*model.AuthTokens, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: login: %w", err))
	}
	if user == nil || auth.VerifyPassword(user.PasswordHash, req.Password) != nil {
		return nil, apperror.Unauthorized("Email atau password salah", ErrInvalidCredentials)
	}

	return s.issueTokenPair(user)
}

// Refresh rotates a refresh token: the presented token is verified,
// rejected if already used, blacklisted so it can't be replayed, and a new
// pair is issued using the user's current roles/permissions.
func (s *AuthService) Refresh(ctx context.Context, req model.RefreshRequest) (*model.AuthTokens, error) {
	claims, err := s.tokens.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, apperror.Unauthorized("Refresh token tidak valid", err)
	}

	blacklisted, err := s.blacklist.IsBlacklisted(ctx, claims.ID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: refresh: %w", err))
	}
	if blacklisted {
		return nil, apperror.Unauthorized("Refresh token sudah tidak berlaku", ErrRefreshTokenReused)
	}

	remaining := time.Until(claims.ExpiresAt.Time)
	if err := s.blacklist.Add(ctx, claims.ID, remaining); err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: refresh: %w", err))
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: refresh: %w", err))
	}
	if user == nil {
		return nil, apperror.Unauthorized("User tidak ditemukan", ErrInvalidCredentials)
	}

	return s.issueTokenPair(user)
}

// Me returns the currently authenticated user's own profile.
func (s *AuthService) Me(ctx context.Context, userID uint) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: me: %w", err))
	}
	if user == nil {
		return nil, apperror.NotFound("User tidak ditemukan", nil)
	}
	return user, nil
}

func (s *AuthService) issueTokenPair(user *model.User) (*model.AuthTokens, error) {
	accessToken, err := s.tokens.IssueAccessToken(user.ID, user.RoleNames(), user.PermissionCodes())
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: issue token pair: %w", err))
	}
	refreshToken, err := s.tokens.IssueRefreshToken(user.ID)
	if err != nil {
		return nil, apperror.Internal(fmt.Errorf("service: issue token pair: %w", err))
	}

	return &model.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.tokens.AccessTTLSeconds(),
	}, nil
}

// validatePasswordStrength enforces mixed case + digit, on top of the
// length check already done by the DTO binding tag (min=8,max=72).
func validatePasswordStrength(password string) error {
	var hasUpper, hasLower, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	if hasUpper && hasLower && hasDigit {
		return nil
	}
	return apperror.Validation("Password harus mengandung huruf besar, huruf kecil, dan angka", []string{
		"password: must contain uppercase, lowercase, and a digit",
	})
}
