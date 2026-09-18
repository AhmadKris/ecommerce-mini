package router

import "github.com/gin-gonic/gin"

// registerAuthRoutes registers the public authentication endpoints.
func registerAuthRoutes(api *gin.RouterGroup, deps Deps) {
	auth := api.Group("/auth")
	auth.POST("/register", deps.AuthHandler.Register)
	auth.POST("/login", deps.AuthHandler.Login)
	auth.POST("/refresh", deps.AuthHandler.Refresh)
}
