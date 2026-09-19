// Package service holds business logic per resource: stock checks at
// checkout, email uniqueness at registration, multi-repository orchestration
// inside transactions. It translates repository errors into
// *apperror.AppError before they reach the handler layer.
package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"
	"unicode"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/repository"
)

// passwordResetTTL is how long a forgot-password token stays redeemable.
// Short on purpose — unlike a session, nothing legitimate needs it to live
// long, and a shorter window shrinks the damage from an intercepted link.
const passwordResetTTL = 30 * time.Minute

// PasswordResetStore issues and redeems single-use password reset tokens.
// Defined here (not imported as a concrete type) so AuthService depends on
// the behavior it needs, not on auth.PasswordResetStore's Redis backing —
// same reasoning as every other repository interface in this codebase.
type PasswordResetStore interface {
	GenerateToken(ctx context.Context, userID uint, ttl time.Duration) (string, error)
	Consume(ctx context.Context, token string) (uint, error)
}

// ErrInvalidCredentials means the email/password pair did not match. It is
// never distinguished by field in the response, so a client can't use the
// error to enumerate registered emails.
var ErrInvalidCredentials = errors.New("invalid email or password")

// ErrRefreshTokenReused means a refresh token was presented after it had
// already been rotated (or otherwise blacklisted).
var ErrRefreshTokenReused = errors.New("refresh token already used or revoked")

const customerRoleName = "customer"

// AuthService implements registration, login, refresh-token rotation, and
// password reset.
type AuthService struct {
	userRepo   repository.UserRepository
	roleRepo   repository.RoleRepository
	tokens     *auth.TokenManager
	blacklist  *auth.RefreshBlacklist
	resetStore PasswordResetStore
	bcryptCost int
}

// NewAuthService builds an AuthService with its dependencies.
func NewAuthService(
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	tokens *auth.TokenManager,
	blacklist *auth.RefreshBlacklist,
	resetStore PasswordResetStore,
	bcryptCost int,
) *AuthService {
	return &AuthService{
		userRepo:   userRepo,
		roleRepo:   roleRepo,
		tokens:     tokens,
		blacklist:  blacklist,
		resetStore: resetStore,
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

// ForgotPassword issues a reset token for email if an account with that
// email exists. It never reports whether the email was found — the handler
// always returns the same generic success message — so this endpoint can't
// be used to enumerate registered accounts.
//
// There's no email provider wired into this project yet (see
// .claude/CLAUDE.md Known Issues): the token is logged instead of sent,
// which is fine for a portfolio/demo but is explicitly not production
// behavior — replace with a real send once a provider is chosen.
func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return apperror.Internal(fmt.Errorf("service: forgot password: %w", err))
	}
	if user == nil {
		return nil
	}

	token, err := s.resetStore.GenerateToken(ctx, user.ID, passwordResetTTL)
	if err != nil {
		return apperror.Internal(fmt.Errorf("service: forgot password: %w", err))
	}

	slog.Info("password reset requested",
		slog.Uint64("user_id", uint64(user.ID)),
		slog.String("email", user.Email),
		slog.String("reset_token", token),
	)
	return nil
}

// ResetPassword redeems token and sets the account it was issued for to
// newPassword. The token is single-use — Consume deletes it on read — so a
// second attempt with the same token fails even if the first succeeded.
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	if err := validatePasswordStrength(newPassword); err != nil {
		return err
	}

	userID, err := s.resetStore.Consume(ctx, token)
	if err != nil {
		return apperror.Internal(fmt.Errorf("service: reset password: %w", err))
	}
	if userID == 0 {
		return apperror.Unauthorized("Token reset password tidak valid atau sudah kedaluwarsa", nil)
	}

	passwordHash, err := auth.HashPassword(newPassword, s.bcryptCost)
	if err != nil {
		return apperror.Internal(fmt.Errorf("service: reset password: %w", err))
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return apperror.Internal(fmt.Errorf("service: reset password: %w", err))
	}
	return nil
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
