package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerCategoryRoutes registers the public category list and the admin
// write endpoints (permission-gated), mirroring product_routes.go.
func registerCategoryRoutes(api *gin.RouterGroup, deps Deps) {
	categories := api.Group("/categories")
	categories.GET("", deps.CategoryHandler.List)

	adminCategories := api.Group("/admin/categories", middleware.RequireAuth(deps.Tokens))
	adminCategories.POST("", middleware.RequirePermission("category:create"), deps.CategoryHandler.Create)
	adminCategories.PUT("/:id", middleware.RequirePermission("category:update"), deps.CategoryHandler.Update)
	adminCategories.DELETE("/:id", middleware.RequirePermission("category:delete"), deps.CategoryHandler.Delete)
}
