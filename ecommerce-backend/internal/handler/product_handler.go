package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
)

// ProductHandler exposes the product catalog: public read, admin write.
type ProductHandler struct {
	productService *service.ProductService
}

// NewProductHandler builds a ProductHandler backed by productService.
func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{productService: productService}
}

// List handles GET /api/products.
func (h *ProductHandler) List(c *gin.Context) {
	var query model.ProductListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(apperror.Validation("Parameter pencarian produk tidak valid", bindingErrors(err)))
		return
	}

	products, meta, err := h.productService.List(c.Request.Context(), query)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"items": products,
			"meta":  meta,
		},
	})
}

// GetBySlug handles GET /api/products/:slug.
func (h *ProductHandler) GetBySlug(c *gin.Context) {
	product, err := h.productService.GetBySlug(c.Request.Context(), c.Param("slug"))
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": product})
}

// Create handles POST /api/admin/products.
func (h *ProductHandler) Create(c *gin.Context) {
	var req model.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data produk tidak valid", bindingErrors(err)))
		return
	}

	product, err := h.productService.Create(c.Request.Context(), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "data": product})
}

// Update handles PUT /api/admin/products/:id.
func (h *ProductHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID produk tidak valid", []string{"id: must be an integer"}))
		return
	}

	var req model.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data produk tidak valid", bindingErrors(err)))
		return
	}

	product, err := h.productService.Update(c.Request.Context(), uint(id), req)
	if err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": product})
}
