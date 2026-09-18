// Package router wires together middleware and route registration for the
// API. Route groups are added here, not in main.go.
package router

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/handler"
	"ecommerce-backend/internal/middleware"
)

// Deps carries the dependencies route handlers need. It is built once in
// cmd/api and passed down instead of using globals.
type Deps struct {
	DB     *gorm.DB
	Cache  *redis.Client
	Logger *slog.Logger

	CORSAllowedOrigins []string

	Tokens           *auth.TokenManager
	AuthHandler      *handler.AuthHandler
	ProductHandler   *handler.ProductHandler
	CategoryHandler  *handler.CategoryHandler
	CartHandler      *handler.CartHandler
	OrderHandler     *handler.OrderHandler
	AddressHandler   *handler.AddressHandler
	InventoryHandler *handler.InventoryHandler
}

// New builds the Gin engine with global middleware and all route groups
// registered.
func New(deps Deps) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.RequestID(deps.Logger))
	engine.Use(middleware.RequestLogging())
	engine.Use(middleware.CORS(deps.CORSAllowedOrigins))
	// ErrorHandler must be registered before anything that can itself call
	// c.Error()+c.Abort() (MaxBodySize here, RequirePermission/RateLimit at
	// the route level) — it works by calling c.Next() and inspecting
	// c.Errors *after* it returns, so it has to wrap around them in the
	// chain, not come after them. Getting this backwards means c.Abort()
	// short-circuits before ErrorHandler's c.Next() is ever reached, and the
	// client gets a bare 200 with no body instead of the intended error
	// response — found by actually curling an oversized request, not by
	// reading the code.
	engine.Use(middleware.ErrorHandler())
	engine.Use(middleware.MaxBodySize())

	engine.GET("/health", healthHandler())
	engine.GET("/ready", readyHandler(deps))

	api := engine.Group("/api")
	registerAuthRoutes(api, deps)
	registerProductRoutes(api, deps)
	registerCategoryRoutes(api, deps)
	registerCartRoutes(api, deps)
	registerOrderRoutes(api, deps)
	registerAddressRoutes(api, deps)
	registerInventoryRoutes(api, deps)

	return engine
}

// healthHandler reports liveness only — no dependency checks — so the
// orchestrator does not restart a healthy process just because Postgres or
// Redis is briefly unavailable.
func healthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true, "data": gin.H{"status": "ok"}})
	}
}

// readyHandler reports readiness to serve traffic, checking that Postgres
// and Redis are actually reachable.
func readyHandler(deps Deps) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := deps.DB.DB()
		if err != nil || sqlDB.PingContext(c.Request.Context()) != nil {
			c.JSON(503, gin.H{"success": false, "error": gin.H{"code": "NOT_READY", "message": "database unavailable"}})
			return
		}

		if err := deps.Cache.Ping(c.Request.Context()).Err(); err != nil {
			c.JSON(503, gin.H{"success": false, "error": gin.H{"code": "NOT_READY", "message": "cache unavailable"}})
			return
		}

		c.JSON(200, gin.H{"success": true, "data": gin.H{"status": "ready"}})
	}
}
