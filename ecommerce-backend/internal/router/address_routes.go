package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerAddressRoutes registers the authenticated user's own address book
// endpoints. No permission gate beyond auth — managing your own addresses
// isn't an admin action, same reasoning as cart and GET /auth/me.
func registerAddressRoutes(api *gin.RouterGroup, deps Deps) {
	addresses := api.Group("/profile/addresses", middleware.RequireAuth(deps.Tokens))
	addresses.GET("", deps.AddressHandler.List)
	addresses.POST("", deps.AddressHandler.Create)
	addresses.PUT("/:id", deps.AddressHandler.Update)
	addresses.DELETE("/:id", deps.AddressHandler.Delete)
}
