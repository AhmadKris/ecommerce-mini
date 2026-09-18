package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/auth"
)

const (
	contextKeyUserID      = "auth.user_id"
	contextKeyPermissions = "auth.permissions"
)

// RequireAuth verifies the Bearer access token and, on success, stores the
// authenticated user's ID and permissions on the gin context for downstream
// handlers/middleware (RequirePermission, ownership checks in service).
func RequireAuth(tokens *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenString, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || tokenString == "" {
			_ = c.Error(apperror.Unauthorized("Token akses tidak ditemukan", nil))
			c.Abort()
			return
		}

		claims, err := tokens.ParseAccessToken(tokenString)
		if err != nil {
			_ = c.Error(apperror.Unauthorized("Token akses tidak valid atau kedaluwarsa", err))
			c.Abort()
			return
		}

		c.Set(contextKeyUserID, claims.UserID)
		c.Set(contextKeyPermissions, claims.Permissions)
		c.Next()
	}
}

// RequirePermission rejects the request with 403 unless the authenticated
// user's token carries the given permission code. It answers "can this
// request reach the endpoint at all?" — row-level ownership (e.g. "is this
// order the caller's own?") is a separate check in the service layer.
func RequirePermission(code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		permissions, _ := c.Get(contextKeyPermissions)
		granted, _ := permissions.([]string)

		for _, p := range granted {
			if p == code {
				c.Next()
				return
			}
		}

		_ = c.Error(apperror.Forbidden("Anda tidak memiliki izin untuk melakukan aksi ini"))
		c.Abort()
	}
}

// UserIDFromContext retrieves the authenticated user's ID set by
// RequireAuth. ok is false if called outside an authenticated request.
func UserIDFromContext(c *gin.Context) (uint, bool) {
	userID, ok := c.Get(contextKeyUserID)
	if !ok {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}
