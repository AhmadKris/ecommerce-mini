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

// CategoryHandler exposes the category catalog: public read, admin write.
type CategoryHandler struct {
	categoryService *service.CategoryService
}

// NewCategoryHandler builds a CategoryHandler backed by categoryService.
func NewCategoryHandler(categoryService *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{categoryService: categoryService}
}

// List handles GET /api/categories.
func (h *CategoryHandler) List(c *gin.Context) {
	categories, err := h.categoryService.List(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": categories}})
}

// Create handles POST /api/admin/categories.
func (h *CategoryHandler) Create(c *gin.Context) {
	var req model.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data kategori tidak valid", bindingErrors(err)))
		return
	}

	actorID, _ := middleware.UserIDFromContext(c)
	category, err := h.categoryService.Create(c.Request.Context(), actorID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": category})
}

// Update handles PUT /api/admin/categories/:id.
func (h *CategoryHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID kategori tidak valid", []string{"id: must be an integer"}))
		return
	}

	var req model.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data kategori tidak valid", bindingErrors(err)))
		return
	}

	actorID, _ := middleware.UserIDFromContext(c)
	category, err := h.categoryService.Update(c.Request.Context(), actorID, uint(id), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": category})
}

// Delete handles DELETE /api/admin/categories/:id.
func (h *CategoryHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID kategori tidak valid", []string{"id: must be an integer"}))
		return
	}

	actorID, _ := middleware.UserIDFromContext(c)
	if err := h.categoryService.Delete(c.Request.Context(), actorID, uint(id)); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"deleted": true}})
}
