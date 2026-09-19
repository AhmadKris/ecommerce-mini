package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerPromotionRoutes registers the admin discount-code CRUD
// endpoints. One permission (promotion:manage) covers all of create/
// update/delete — unlike products/categories, there's no independent use
// case yet for "can create promos but not delete them", so splitting into
// 3 permissions now would be speculative (see §12 no premature abstraction).
func registerPromotionRoutes(api *gin.RouterGroup, deps Deps) {
	promotions := api.Group("/admin/promotions", middleware.RequireAuth(deps.Tokens), middleware.RequirePermission("promotion:manage"))
	promotions.GET("", deps.PromotionHandler.List)
	promotions.POST("", deps.PromotionHandler.Create)
	promotions.PUT("/:id", deps.PromotionHandler.Update)
	promotions.DELETE("/:id", deps.PromotionHandler.Delete)
}
