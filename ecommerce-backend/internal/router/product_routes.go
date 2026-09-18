package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerProductRoutes registers the public product catalog endpoints and
// the admin write endpoints (permission-gated).
func registerProductRoutes(api *gin.RouterGroup, deps Deps) {
	products := api.Group("/products")
	products.GET("", deps.ProductHandler.List)
	products.GET("/:slug", deps.ProductHandler.GetBySlug)

	adminProducts := api.Group("/admin/products", middleware.RequireAuth(deps.Tokens))
	adminProducts.POST("", middleware.RequirePermission("product:create"), deps.ProductHandler.Create)
	adminProducts.PUT("/:id", middleware.RequirePermission("product:update"), deps.ProductHandler.Update)
	adminProducts.DELETE("/:id", middleware.RequirePermission("product:delete"), deps.ProductHandler.Delete)
}
