package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
)

// CartHandler exposes the authenticated user's own cart. Ownership (which
// cart) always comes from the JWT via middleware.UserIDFromContext, never
// from a request parameter — there is no "whose cart" to bind from the
// client.
type CartHandler struct {
	cartService *service.CartService
}

// NewCartHandler builds a CartHandler backed by cartService.
func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

// Get handles GET /api/cart.
func (h *CartHandler) Get(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)
	cart, err := h.cartService.Get(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cart})
}

// AddItem handles POST /api/cart/items.
func (h *CartHandler) AddItem(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	var req model.AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data item cart tidak valid", bindingErrors(err)))
		return
	}

	cart, err := h.cartService.AddItem(c.Request.Context(), userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": cart})
}

// UpdateItem handles PUT /api/cart/items/:id.
func (h *CartHandler) UpdateItem(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID item cart tidak valid", []string{"id: must be an integer"}))
		return
	}

	var req model.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data item cart tidak valid", bindingErrors(err)))
		return
	}

	cart, err := h.cartService.UpdateItemQuantity(c.Request.Context(), userID, uint(itemID), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cart})
}

// RemoveItem handles DELETE /api/cart/items/:id.
func (h *CartHandler) RemoveItem(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	itemID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID item cart tidak valid", []string{"id: must be an integer"}))
		return
	}

	cart, err := h.cartService.RemoveItem(c.Request.Context(), userID, uint(itemID))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": cart})
}
