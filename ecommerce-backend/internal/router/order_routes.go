package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerOrderRoutes registers the authenticated checkout/order-history
// endpoints. Checkout itself needs only authentication (any logged-in
// customer can check out their own cart); listing is additionally gated on
// order:read_own since it's a read-scope permission by convention (see
// product/cart routes).
func registerOrderRoutes(api *gin.RouterGroup, deps Deps) {
	orders := api.Group("/orders", middleware.RequireAuth(deps.Tokens))
	orders.POST("", deps.OrderHandler.Checkout)
	orders.GET("", middleware.RequirePermission("order:read_own"), deps.OrderHandler.List)
}
