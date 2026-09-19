package router

import (
	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// registerReviewRoutes registers the public review read/create endpoints
// (create requires auth — a review needs a real, owned delivered order) and
// the admin moderation endpoint.
func registerReviewRoutes(api *gin.RouterGroup, deps Deps) {
	productReviews := api.Group("/products/:slug/reviews")
	productReviews.GET("", deps.ReviewHandler.ListForProduct)
	productReviews.POST("", middleware.RequireAuth(deps.Tokens), deps.ReviewHandler.Create)

	adminReviews := api.Group("/admin/reviews", middleware.RequireAuth(deps.Tokens), middleware.RequirePermission("review:moderate"))
	adminReviews.GET("", deps.ReviewHandler.ListAll)
	adminReviews.PATCH("/:id/status", deps.ReviewHandler.UpdateStatus)
}
