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

// CustomerHandler exposes admin read access to registered accounts and role
// management.
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

// ListRoles handles GET /api/admin/roles — returns all available roles with
// their permissions, so the admin UI can populate the role selection UI.
func (h *CustomerHandler) ListRoles(c *gin.Context) {
	roles, err := h.customerService.ListRoles(c.Request.Context())
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": roles})
}

// UpdateUserRoles handles PUT /api/admin/customers/:id/roles — replaces the
// user's full role set with the submitted list.
func (h *CustomerHandler) UpdateUserRoles(c *gin.Context) {
	targetID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID customer tidak valid", []string{"id: must be an integer"}))
		return
	}

	var req model.UpdateUserRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Request body tidak valid", bindingErrors(err)))
		return
	}

	callerID, ok := middleware.UserIDFromContext(c)
	if !ok {
		_ = c.Error(apperror.Unauthorized("Token tidak valid", nil))
		return
	}

	if err := h.customerService.UpdateUserRoles(c.Request.Context(), callerID, uint(targetID), req.RoleNames); err != nil {
		_ = c.Error(err)
		return
	}

	// Fetch the updated user so the response reflects the new state without
	// requiring a separate client round-trip.
	updated, err := h.customerService.GetByID(c.Request.Context(), uint(targetID))
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": updated, "message": "Role berhasil diperbarui"})
}
