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

// AddressHandler exposes the authenticated user's own address book.
type AddressHandler struct {
	addressService *service.AddressService
}

// NewAddressHandler builds an AddressHandler backed by addressService.
func NewAddressHandler(addressService *service.AddressService) *AddressHandler {
	return &AddressHandler{addressService: addressService}
}

// List handles GET /api/profile/addresses.
func (h *AddressHandler) List(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	addresses, err := h.addressService.List(c.Request.Context(), userID)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"items": addresses}})
}

// Create handles POST /api/profile/addresses.
func (h *AddressHandler) Create(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	var req model.CreateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data alamat tidak valid", bindingErrors(err)))
		return
	}

	address, err := h.addressService.Create(c.Request.Context(), userID, req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": true, "data": address})
}

// Update handles PUT /api/profile/addresses/:id.
func (h *AddressHandler) Update(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID alamat tidak valid", []string{"id: must be an integer"}))
		return
	}

	var req model.UpdateAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperror.Validation("Data alamat tidak valid", bindingErrors(err)))
		return
	}

	address, err := h.addressService.Update(c.Request.Context(), userID, uint(id), req)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": address})
}

// Delete handles DELETE /api/profile/addresses/:id.
func (h *AddressHandler) Delete(c *gin.Context) {
	userID, _ := middleware.UserIDFromContext(c)

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		_ = c.Error(apperror.Validation("ID alamat tidak valid", []string{"id: must be an integer"}))
		return
	}

	if err := h.addressService.Delete(c.Request.Context(), userID, uint(id)); err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"deleted": true}})
}
