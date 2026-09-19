package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerCustomerRoutes registers the admin's read-only view of registered
// accounts. One permission (customer:read) covers list and detail — no
// separate "manage" permission yet, since there's no mutation endpoint
// (role/permission editing) to gate; add one only when that need is real.
func registerCustomerRoutes(api *gin.RouterGroup, deps Deps) {
	customers := api.Group("/admin/customers", middleware.RequireAuth(deps.Tokens), middleware.RequirePermission("customer:read"))
	customers.GET("", deps.CustomerHandler.List)
	customers.GET("/:id", deps.CustomerHandler.GetByID)
}
