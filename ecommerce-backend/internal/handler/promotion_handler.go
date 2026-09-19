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

// PromotionHandler exposes discount-code CRUD (admin only).
type PromotionHandler struct {
	promotionService *service.PromotionService
}

// NewPromotionHandler builds a PromotionHandler backed by promotionService.
func NewPromotionHandler(promotionService *service.PromotionService) *PromotionHandler {
	return &PromotionHandler{promotionService: promotionService}
}

// List handles GET /api/admin/promotions.
func (h *PromotionHandler) List(c *gin.Context) {
	promotions, err := h.promotionService.List(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": promotions}})
}

// Create handles POST /api/admin/promotions.
func (h *PromotionHandler) Create(c *gin.Context) {
	var req model.CreatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data promo tidak valid", bindingErrors(err)))
		return
	}

	actorID, _ := middleware.UserIDFromContext(c)
	promotion, err := h.promotionService.Create(c.Request.Context(), actorID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": promotion})
}

// Update handles PUT /api/admin/promotions/:id.
func (h *PromotionHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID promo tidak valid", []string{"id: must be an integer"}))
		return
	}

	var req model.UpdatePromotionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data promo tidak valid", bindingErrors(err)))
		return
	}

	actorID, _ := middleware.UserIDFromContext(c)
	promotion, err := h.promotionService.Update(c.Request.Context(), actorID, uint(id), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": promotion})
}

// Delete handles DELETE /api/admin/promotions/:id.
func (h *PromotionHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID promo tidak valid", []string{"id: must be an integer"}))
		return
	}

	actorID, _ := middleware.UserIDFromContext(c)
	if err := h.promotionService.Delete(c.Request.Context(), actorID, uint(id)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"deleted": true}})
}
