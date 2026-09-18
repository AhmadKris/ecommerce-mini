package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/middleware"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
)

// OrderHandler exposes checkout and the authenticated user's own order
// history.
type OrderHandler struct {
	orderService *service.OrderService
}

// NewOrderHandler builds an OrderHandler backed by orderService.
func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// Checkout handles POST /api/orders.
func (h *OrderHandler) Checkout(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	var req model.CheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data checkout tidak valid", bindingErrors(err)))
		return
	}

	order, err := h.orderService.Checkout(c.Request.Context(), userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": order})
}

// List handles GET /api/orders.
func (h *OrderHandler) List(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	var query model.OrderListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(apperror.Validation("Parameter pencarian order tidak valid", bindingErrors(err)))
		return
	}

	orders, meta, err := h.orderService.List(c.Request.Context(), userID, query)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": orders, "meta": meta}})
}
