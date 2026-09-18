package auth

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// AccessClaims carries everything an authorization check needs so it never
// has to query the database: user identity plus the roles/permissions
// snapshotted at token-issue time.
type AccessClaims struct {
	jwt.RegisteredClaims
	UserID      uint     `json:"user_id"`
	Roles       []string `json:"roles"`
	Permissions []string `json:"permissions"`
}

// RefreshClaims is intentionally minimal — just enough to identify the user
// and the token (via jti, in RegisteredClaims.ID) for blacklisting. Roles
// and permissions are re-read from the database on refresh instead of being
// carried here, so a revoked role takes effect at the next refresh rather
// than waiting for the access token to expire naturally.
type RefreshClaims struct {
	jwt.RegisteredClaims
	UserID uint `json:"user_id"`
}

// TokenManager issues and verifies the access/refresh JWT pair.
type TokenManager struct {
	accessSecret  []byte
	refreshSecret []byte
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewTokenManager builds a TokenManager. accessSecret and refreshSecret must
// be distinct so a leaked access token can't be replayed as a refresh token.
func NewTokenManager(accessSecret, refreshSecret string, accessTTL, refreshTTL time.Duration) *TokenManager {
	return &TokenManager{
		accessSecret:  []byte(accessSecret),
		refreshSecret: []byte(refreshSecret),
		accessTTL:     accessTTL,
		refreshTTL:    refreshTTL,
	}
}

// AccessTTLSeconds reports the access token lifetime in seconds, for the
// `expires_in` field in the login/refresh response.
func (m *TokenManager) AccessTTLSeconds() int64 {
	return int64(m.accessTTL.Seconds())
}

// IssueAccessToken signs a short-lived access token carrying userID's roles
// and permissions.
func (m *TokenManager) IssueAccessToken(userID uint, roles, permissions []string) (string, error) {
	now := time.Now()
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(uint64(userID), 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
		UserID:      userID,
		Roles:       roles,
		Permissions: permissions,
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.accessSecret)
	if err != nil {
		return "", fmt.Errorf("auth: issue access token: %w", err)
	}
	return signed, nil
}

// IssueRefreshToken signs a longer-lived refresh token with a fresh jti, so
// it can be individually blacklisted on rotation.
func (m *TokenManager) IssueRefreshToken(userID uint) (string, error) {
	now := time.Now()
	claims := RefreshClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        uuid.NewString(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTTL)),
		},
		UserID: userID,
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.refreshSecret)
	if err != nil {
		return "", fmt.Errorf("auth: issue refresh token: %w", err)
	}
	return signed, nil
}

// ParseAccessToken verifies signature and expiry and returns the claims.
func (m *TokenManager) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return m.accessSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("auth: parse access token: %w", err)
	}
	return claims, nil
}

// ParseRefreshToken verifies signature and expiry and returns the claims.
func (m *TokenManager) ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	claims := &RefreshClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return m.refreshSecret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("auth: parse refresh token: %w", err)
	}
	return claims, nil
}
