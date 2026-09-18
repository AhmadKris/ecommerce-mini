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

// InventoryHandler exposes stock adjustments and per-product movement
// history (admin only).
type InventoryHandler struct {
	inventoryService *service.InventoryService
}

// NewInventoryHandler builds an InventoryHandler backed by inventoryService.
func NewInventoryHandler(inventoryService *service.InventoryService) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService}
}

// Adjust handles POST /api/admin/inventory/adjustments.
func (h *InventoryHandler) Adjust(c *gin.Context) {
	var req model.AdjustInventoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data penyesuaian stok tidak valid", bindingErrors(err)))
		return
	}

	actorID, _ := middleware.UserIDFromContext(c)
	movement, err := h.inventoryService.Adjust(c.Request.Context(), actorID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": movement})
}

// ListMovements handles GET /api/admin/inventory/:productId/movements.
func (h *InventoryHandler) ListMovements(c *gin.Context) {
	productID, err := strconv.ParseUint(c.Param("productId"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID produk tidak valid", []string{"productId: must be an integer"}))
		return
	}

	var query model.InventoryMovementListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(apperror.Validation("Parameter pencarian tidak valid", bindingErrors(err)))
		return
	}

	movements, meta, err := h.inventoryService.ListMovements(c.Request.Context(), uint(productID), query)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": movements, "meta": meta}})
}
