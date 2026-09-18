package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
)

// AuthHandler exposes registration, login, and refresh-token rotation.
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler builds an AuthHandler backed by authService.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data registrasi tidak valid", bindingErrors(err)))
		return
	}

	user, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    model.ToUserResponse(user),
	})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data login tidak valid", bindingErrors(err)))
		return
	}

	tokens, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tokens,
	})
}

// Me handles GET /api/auth/me — the authenticated caller's own profile.
func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	user, err := h.authService.Me(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    model.ToUserResponse(user),
	})
}

// Refresh handles POST /api/auth/refresh.
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req model.RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data refresh token tidak valid", bindingErrors(err)))
		return
	}

	tokens, err := h.authService.Refresh(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tokens,
	})
}
