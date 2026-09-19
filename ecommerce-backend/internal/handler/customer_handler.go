package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/apperror"
	"ecommerce-backend/internal/model"
	"ecommerce-backend/internal/service"
)

// CustomerHandler exposes admin read access to registered accounts.
type CustomerHandler struct {
	customerService *service.CustomerService
}

// NewCustomerHandler builds a CustomerHandler backed by customerService.
func NewCustomerHandler(customerService *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{customerService: customerService}
}

// List handles GET /api/admin/customers.
func (h *CustomerHandler) List(c *gin.Context) {
	var query model.CustomerListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		_ = c.Error(apperror.Validation("Parameter pencarian tidak valid", bindingErrors(err)))
		return
	}

	customers, meta, err := h.customerService.List(c.Request.Context(), query)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": customers, "meta": meta}})
}

// GetByID handles GET /api/admin/customers/:id.
func (h *CustomerHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID customer tidak valid", []string{"id: must be an integer"}))
		return
	}

	customer, err := h.customerService.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": customer})
}
