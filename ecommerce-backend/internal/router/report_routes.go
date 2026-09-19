package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerReportRoutes registers the admin business-reports endpoint.
func registerReportRoutes(api *gin.RouterGroup, deps Deps) {
	reports := api.Group("/admin/reports", middleware.RequireAuth(deps.Tokens), middleware.RequirePermission("report:read"))
	reports.GET("/dashboard", deps.ReportHandler.Dashboard)
}
