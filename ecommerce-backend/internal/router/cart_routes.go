package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerCartRoutes registers the authenticated cart endpoints.
func registerCartRoutes(api *gin.RouterGroup, deps Deps) {
	cart := api.Group("/cart", middleware.RequireAuth(deps.Tokens), middleware.RequirePermission("cart:manage"))
	cart.GET("", deps.CartHandler.Get)
	cart.POST("/items", deps.CartHandler.AddItem)
	cart.PUT("/items/:id", deps.CartHandler.UpdateItem)
	cart.DELETE("/items/:id", deps.CartHandler.RemoveItem)
}
