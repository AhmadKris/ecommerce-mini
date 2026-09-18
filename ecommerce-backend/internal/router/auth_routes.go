package router

import (
	"time"

	"github.com/gin-gonic/gin"

	"ecommerce-backend/internal/middleware"
)

// authRateLimit caps each IP at 10 requests/minute across all auth
// endpoints — strict on purpose (PLANNING.md §2A calls out auth
// specifically), since register/login/refresh are exactly what a
// credential-stuffing or brute-force script would hammer.
const authRateLimit = 10

// registerAuthRoutes registers the public authentication endpoints.
func registerAuthRoutes(api *gin.RouterGroup, deps Deps) {
	auth := api.Group("/auth", middleware.RateLimit(deps.Cache, "auth", authRateLimit, time.Minute))
	auth.POST("/register", deps.AuthHandler.Register)
	auth.POST("/login", deps.AuthHandler.Login)
	auth.POST("/refresh", deps.AuthHandler.Refresh)
	auth.GET("/me", middleware.RequireAuth(deps.Tokens), deps.AuthHandler.Me)
}
