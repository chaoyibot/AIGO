package main

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/aigo/internal/api/handler"
	"github.com/aigo/internal/api/middleware"
	"github.com/aigo/internal/api/response"
	"github.com/aigo/internal/config"
	authService "github.com/aigo/internal/service/auth"
	broadcastService "github.com/aigo/internal/service/broadcast"
	productService "github.com/aigo/internal/service/product"
	tradingService "github.com/aigo/internal/service/trading"
	walletService "github.com/aigo/internal/service/wallet"
	"github.com/aigo/internal/repository"
	"github.com/rs/zerolog"
)

func main() {
	cfg := config.Load()

	// Logger
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).With().Timestamp().Logger()

	// Database
	db, err := repository.Connect(cfg.Database.URL)
	if err != nil {
		logger.Fatal().Err(err).Msg("database connection failed")
	}
	defer db.Close()
	logger.Info().Msg("database connected")

	// Repositories
	userRepo := repository.NewUserRepo(db)
	productRepo := repository.NewProductRepo(db)
	listingRepo := repository.NewListingRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	walletRepo := repository.NewWalletRepo(db)
	rechargeRepo := repository.NewRechargeRepo(db)
	withdrawRepo := repository.NewWithdrawalRepo(db)
	broadcastRepo := repository.NewBroadcastRepo(db)

	// Services
	authSvc := authService.NewService(userRepo, "aigo-jwt-secret-change-in-production")
	productSvc := productService.NewService(productRepo)
	tradingSvc := tradingService.NewService(db, productRepo, listingRepo, walletRepo)
	walletSvc := walletService.NewService(walletRepo, rechargeRepo, withdrawRepo, cfg, db)
	broadcastSvc := broadcastService.NewService(broadcastRepo, walletRepo, cfg, db)

	// Handlers
	authHandler := handler.NewAuthHandler(authSvc)
	productHandler := handler.NewProductHandler(productSvc)
	listingHandler := handler.NewListingHandler(listingRepo, tradingSvc)
	orderHandler := handler.NewOrderHandler(orderRepo)
	walletHandler := handler.NewWalletHandler(walletSvc)
	broadcastHandler := handler.NewBroadcastHandler(broadcastSvc)
	webhookHandler := handler.NewWebhookHandler()
	eventHandler := handler.NewEventHandler()

	// Router
	r := gin.Default()
	r.Use(gin.Recovery())

	// Request ID middleware
	r.Use(func(c *gin.Context) {
		c.Set("request_id", generateRequestID())
		c.Next()
	})

	// Health
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok", "version": "0.2.0", "name": "AIGO"})
	})

	// API v1
	v1 := r.Group("/api/v1")
	{
		// Auth (no auth required)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/token", authHandler.ExchangeToken)
		}

		// Authenticated routes
		authed := v1.Group("")
		authed.Use(middleware.AuthRequired(authSvc))
		{
			// API Key management
			authed.POST("/auth/api-key", authHandler.GenerateAPIKey)

			// Products
			authed.GET("/products", productHandler.List)
			authed.POST("/products", productHandler.Create)
			authed.GET("/products/:id", productHandler.GetByID)
			authed.PUT("/products/:id", productHandler.Update)
			authed.DELETE("/products/:id", productHandler.Delete)

			// Listings
			authed.GET("/listings", listingHandler.List)
			authed.POST("/listings", listingHandler.Create)
			authed.POST("/listings/:id/buy", listingHandler.Buy)

			// Orders
			authed.GET("/orders", orderHandler.List)
			authed.GET("/orders/:id", orderHandler.GetByID)
			authed.POST("/orders/:id/confirm", orderHandler.Confirm)
			authed.POST("/orders/:id/cancel", orderHandler.Cancel)
			authed.POST("/orders/:id/dispute", orderHandler.Dispute)

			// Wallet & Recharge
			authed.GET("/wallet", walletHandler.GetBalance)
			authed.POST("/wallet/convert", walletHandler.Convert)
			authed.GET("/wallet/transactions", walletHandler.GetTransactions)
			authed.POST("/recharge", walletHandler.Recharge)

			// Broadcasts
			authed.GET("/broadcasts", broadcastHandler.List)
			authed.GET("/broadcasts/unread", broadcastHandler.UnreadCount)
			authed.POST("/broadcasts/commercial", broadcastHandler.CommercialBroadcast)
			authed.PUT("/broadcasts/:id/read", broadcastHandler.MarkRead)

			// Webhooks
			authed.POST("/webhooks", webhookHandler.Register)
			authed.GET("/webhooks", webhookHandler.List)
			authed.PUT("/webhooks/:id", webhookHandler.Update)
			authed.DELETE("/webhooks/:id", webhookHandler.Delete)
			authed.GET("/webhooks/:id/logs", webhookHandler.Logs)

			// SSE events
			authed.GET("/events", eventHandler.Stream)
		}
	}

	// Admin routes (system broadcast)
	admin := v1.Group("/admin")
	admin.Use(middleware.AuthRequired(authSvc))
	{
		admin.POST("/broadcasts/system", broadcastHandler.SystemBroadcast)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info().Str("port", port).Msg("AIGO server starting")
	if err := r.Run(":" + port); err != nil {
		logger.Fatal().Err(err).Msg("server failed")
	}
}

func generateRequestID() string {
	b := make([]byte, 16)
	for i := 0; i < 8; i++ {
		n := time.Now().Nanosecond() % 36
		if n < 10 {
			b[i] = byte('0' + n)
		} else {
			b[i] = byte('a' + n - 10)
		}
	}
	b[8] = '-'
	for i := 9; i < 16; i++ {
		n := time.Now().UnixMilli() % 36
		if n < 10 {
			b[i] = byte('0' + n)
		} else {
			b[i] = byte('a' + n - 10)
		}
	}
	return "req_" + string(b)
}
