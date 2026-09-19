// Command api is the entry point for the e-commerce REST API. It wires
// config, logger, database, cache, and router together, then starts the
// HTTP server with graceful shutdown on SIGTERM/SIGINT.
package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ecommerce-backend/internal/auth"
	"ecommerce-backend/internal/cache"
	"ecommerce-backend/internal/config"
	"ecommerce-backend/internal/database"
	"ecommerce-backend/internal/handler"
	"ecommerce-backend/internal/logger"
	"ecommerce-backend/internal/repository"
	"ecommerce-backend/internal/router"
	"ecommerce-backend/internal/service"
	"ecommerce-backend/internal/worker"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.Any("err", err))
		os.Exit(1)
	}

	log := logger.New(cfg.Env)
	slog.SetDefault(log)

	db, err := database.Connect(database.Options{
		DSN:             cfg.DatabaseURL,
		MaxOpenConns:    cfg.DBMaxOpenConns,
		MaxIdleConns:    cfg.DBMaxIdleConns,
		ConnMaxLifetime: cfg.DBConnMaxLifetime,
	})
	if err != nil {
		log.Error("failed to connect to database", slog.Any("err", err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	redisClient, err := cache.Connect(ctx, cache.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	if err != nil {
		log.Error("failed to connect to redis", slog.Any("err", err))
		os.Exit(1)
	}

	tokenManager := auth.NewTokenManager(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	refreshBlacklist := auth.NewRefreshBlacklist(redisClient)
	passwordResetStore := auth.NewPasswordResetStore(redisClient)

	cacheService := cache.NewCacheService(redisClient)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	authService := service.NewAuthService(userRepo, roleRepo, tokenManager, refreshBlacklist, passwordResetStore, cfg.BcryptCost)
	authHandler := handler.NewAuthHandler(authService)

	auditLogRepo := repository.NewAuditLogRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)

	outboxWorker := worker.NewOutboxWorker(outboxRepo, 2*time.Second)
	go outboxWorker.Start(ctx)

	productRepo := repository.NewProductRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	productService := service.NewProductService(productRepo, categoryRepo, auditLogRepo)
	productService.SetCacheService(cacheService)
	productHandler := handler.NewProductHandler(productService)

	categoryService := service.NewCategoryService(categoryRepo, auditLogRepo)
	categoryService.SetCacheService(cacheService)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	cartRepo := repository.NewCartRepository(db)
	cartService := service.NewCartService(cartRepo, productRepo)
	cartHandler := handler.NewCartHandler(cartService)

	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, auditLogRepo)
	orderHandler := handler.NewOrderHandler(orderService)

	addressRepo := repository.NewAddressRepository(db)
	addressService := service.NewAddressService(addressRepo)
	addressHandler := handler.NewAddressHandler(addressService)

	inventoryRepo := repository.NewInventoryRepository(db)
	inventoryService := service.NewInventoryService(inventoryRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)

	promotionRepo := repository.NewPromotionRepository(db)
	promotionService := service.NewPromotionService(promotionRepo, auditLogRepo)
	promotionHandler := handler.NewPromotionHandler(promotionService)

	reviewRepo := repository.NewReviewRepository(db)
	reviewService := service.NewReviewService(reviewRepo, productRepo, auditLogRepo)
	reviewHandler := handler.NewReviewHandler(reviewService)

	customerService := service.NewCustomerService(userRepo, roleRepo, auditLogRepo)
	customerHandler := handler.NewCustomerHandler(customerService)

	reportRepo := repository.NewReportRepository(db)
	reportService := service.NewReportService(reportRepo)
	reportHandler := handler.NewReportHandler(reportService)

	engine := router.New(router.Deps{
		DB:                 db,
		Cache:              redisClient,
		Logger:             log,
		CORSAllowedOrigins: cfg.CORSAllowedOrigins,
		Tokens:             tokenManager,
		AuthHandler:        authHandler,
		ProductHandler:     productHandler,
		CategoryHandler:    categoryHandler,
		CartHandler:        cartHandler,
		OrderHandler:       orderHandler,
		AddressHandler:     addressHandler,
		InventoryHandler:   inventoryHandler,
		PromotionHandler:   promotionHandler,
		ReviewHandler:      reviewHandler,
		CustomerHandler:    customerHandler,
		ReportHandler:      reportHandler,
	})

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: engine,
		// Without this, a client that trickles request headers in slowly
		// (deliberately or not) can hold a connection open indefinitely —
		// the classic Slowloris DoS. 5s is generous for any legitimate
		// client on any network.
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		log.Info("starting server", slog.String("port", cfg.Port), slog.String("env", cfg.Env))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server error", slog.Any("err", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	log.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("graceful shutdown failed", slog.Any("err", err))
		os.Exit(1)
	}

	log.Info("server stopped")
}
