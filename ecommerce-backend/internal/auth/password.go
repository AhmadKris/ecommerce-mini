// Package auth implements the authentication mechanics shared by the auth
// service: password hashing, JWT issuance/verification, and refresh-token
// blacklisting. Business rules (password strength, credential checks) live
// in internal/service — this package only handles the crypto/token plumbing.
package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword hashes password with bcrypt at the given cost. cost is passed
// in explicitly (config.BcryptCost) rather than using bcrypt.DefaultCost, so
// the time-per-hash tradeoff is a deliberate, documented config value.
func HashPassword(password string, cost int) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	if err != nil {
		return "", fmt.Errorf("auth: hash password: %w", err)
	}
	return string(hash), nil
}

// VerifyPassword returns nil if password matches hash, or bcrypt's mismatch
// error otherwise. Callers should only branch on err == nil vs err != nil,
// not on the error's message.
func VerifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
