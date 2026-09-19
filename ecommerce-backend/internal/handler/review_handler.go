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

// ReviewHandler exposes product reviews: public create/read, admin moderation.
type ReviewHandler struct {
	reviewService *service.ReviewService
}

// NewReviewHandler builds a ReviewHandler backed by reviewService.
func NewReviewHandler(reviewService *service.ReviewService) *ReviewHandler {
	return &ReviewHandler{reviewService: reviewService}
}

// Create handles POST /api/products/:slug/reviews.
func (h *ReviewHandler) Create(c *gin.Context) {
	var req model.CreateReviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data ulasan tidak valid", bindingErrors(err)))
		return
	}

	userID, _ := middleware.UserIDFromContext(c)
	review, err := h.reviewService.Create(c.Request.Context(), userID, c.Param("slug"), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": review})
}

// ListForProduct handles GET /api/products/:slug/reviews.
func (h *ReviewHandler) ListForProduct(c *gin.Context) {
	var query model.ReviewListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(apperror.Validation("Parameter tidak valid", bindingErrors(err)))
		return
	}

	reviews, meta, err := h.reviewService.ListApprovedForProduct(c.Request.Context(), c.Param("slug"), query)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": reviews, "meta": meta}})
}

// ListAll handles GET /api/admin/reviews.
func (h *ReviewHandler) ListAll(c *gin.Context) {
	var query model.ReviewListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(apperror.Validation("Parameter tidak valid", bindingErrors(err)))
		return
	}

	reviews, meta, err := h.reviewService.ListAll(c.Request.Context(), query)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": reviews, "meta": meta}})
}

// UpdateStatus handles PATCH /api/admin/reviews/:id/status.
func (h *ReviewHandler) UpdateStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID ulasan tidak valid", []string{"id: must be an integer"}))
		return
	}

	var req model.UpdateReviewStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data tidak valid", bindingErrors(err)))
		return
	}

	actorID, _ := middleware.UserIDFromContext(c)
	review, err := h.reviewService.UpdateStatus(c.Request.Context(), actorID, uint(id), req.Status)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": review})
}
