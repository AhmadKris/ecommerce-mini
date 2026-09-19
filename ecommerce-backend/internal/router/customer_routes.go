package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerCustomerRoutes registers admin endpoints for viewing and managing
// registered customer accounts. customer:read gates the read-only list/detail
// views; user:manage gates the role mutation endpoint — separate permissions
// so a future "support" role could see customers without being able to change
// their access level.
func registerCustomerRoutes(api *gin.RouterGroup, deps Deps) {
	readOnly := api.Group("/admin/customers", middleware.RequireAuth(deps.Tokens), middleware.RequirePermission("customer:read"))
	readOnly.GET("", deps.CustomerHandler.List)
	readOnly.GET("/:id", deps.CustomerHandler.GetByID)

	manage := api.Group("/admin", middleware.RequireAuth(deps.Tokens), middleware.RequirePermission("user:manage"))
	manage.GET("/roles", deps.CustomerHandler.ListRoles)
	manage.PUT("/customers/:id/roles", deps.CustomerHandler.UpdateUserRoles)
}
