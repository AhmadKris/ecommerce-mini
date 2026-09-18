package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerInventoryRoutes registers the admin stock-adjustment endpoints.
func registerInventoryRoutes(api *gin.RouterGroup, deps Deps) {
	inventory := api.Group("/admin/inventory", middleware.RequireAuth(deps.Tokens))
	inventory.POST("/adjustments", middleware.RequirePermission("inventory:adjust"), deps.InventoryHandler.Adjust)
	inventory.GET("/:productId/movements", middleware.RequirePermission("inventory:read"), deps.InventoryHandler.ListMovements)
}
