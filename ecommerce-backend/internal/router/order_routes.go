package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// idempotencyTTL is how long a checkout's Idempotency-Key response stays
// replayable. Generous on purpose — the failure mode it protects against
// (client retries after a dropped connection, possibly minutes later) isn't
// bounded to seconds.
const idempotencyTTL = 24 * time.Hour

// registerOrderRoutes registers the authenticated checkout/order-history
// endpoints. Checkout itself needs only authentication (any logged-in
// customer can check out their own cart); listing is additionally gated on
// order:read_own since it's a read-scope permission by convention (see
// product/cart routes).
func registerOrderRoutes(api *gin.RouterGroup, deps Deps) {
	orders := api.Group("/orders", middleware.RequireAuth(deps.Tokens))
	orders.POST("", middleware.Idempotency(deps.Cache, idempotencyTTL), deps.OrderHandler.Checkout)
	orders.GET("", middleware.RequirePermission("order:read_own"), deps.OrderHandler.List)

	adminOrders := api.Group("/admin/orders", middleware.RequireAuth(deps.Tokens))
	adminOrders.GET("", middleware.RequirePermission("order:read_all"), deps.OrderHandler.ListAll)
	adminOrders.GET("/:id", middleware.RequirePermission("order:read_all"), deps.OrderHandler.GetByID)
	adminOrders.PATCH("/:id/status", middleware.RequirePermission("order:update_status"), deps.OrderHandler.UpdateStatus)
}
