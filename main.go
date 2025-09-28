package main

import (
	"log"
	"os"

	"cursor-ai-backend/docs"
	"cursor-ai-backend/internal/config"
	"cursor-ai-backend/internal/database"
	"cursor-ai-backend/internal/handlers"
	"cursor-ai-backend/internal/middleware"
	"cursor-ai-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
)

// @title License Management System API
// @version 1.0
// @description A comprehensive license management system with admin and customer portals, plus SDK integration
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key
// @description API Key for SDK authentication

func main() {
	// Initialize Swagger docs
	docs.SwaggerInfo.Title = "License Management System API"
	docs.SwaggerInfo.Description = "A comprehensive license management system with admin and customer portals, plus SDK integration"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// Auto-migrate models
	err = db.AutoMigrate(
		&models.User{},
		&models.Customer{},
		&models.SubscriptionPack{},
		&models.Subscription{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Create default admin user if it doesn't exist
	createDefaultAdmin(db)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db)
	customerHandler := handlers.NewCustomerHandler(db)
	packHandler := handlers.NewSubscriptionPackHandler(db)
	subscriptionHandler := handlers.NewSubscriptionHandler(db)
	sdkHandler := handlers.NewSDKHandler(db)

	// Setup router
	router := setupRouter(db, userHandler, customerHandler, packHandler, subscriptionHandler, sdkHandler)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func setupRouter(
	db *database.DB,
	userHandler *handlers.UserHandler,
	customerHandler *handlers.CustomerHandler,
	packHandler *handlers.SubscriptionPackHandler,
	subscriptionHandler *handlers.SubscriptionHandler,
	sdkHandler *handlers.SDKHandler,
) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-API-Key")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Database middleware
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Frontend API routes (JWT authentication)
	api := router.Group("/api")
	{
		// Public authentication endpoints
		auth := api.Group("/")
		{
			auth.POST("/admin/login", userHandler.AdminLogin)
			auth.POST("/customer/login", userHandler.CustomerLogin)
			auth.POST("/customer/signup", userHandler.CustomerSignup)
		}

		// Protected endpoints (JWT required)
		v1 := api.Group("/v1")
		v1.Use(middleware.JWTAuth())
		{
			// Admin-only endpoints
			admin := v1.Group("/admin")
			admin.Use(middleware.AdminOnly())
			{
				// Customer management
				admin.GET("/customers", customerHandler.ListCustomers)
				admin.POST("/customers", customerHandler.CreateCustomer)
				admin.GET("/customers/:id", customerHandler.GetCustomer)
				admin.PUT("/customers/:id", customerHandler.UpdateCustomer)
				admin.DELETE("/customers/:id", customerHandler.DeleteCustomer)

				// Subscription pack management
				admin.GET("/packs", packHandler.ListPacks)
				admin.POST("/packs", packHandler.CreatePack)
				admin.GET("/packs/:id", packHandler.GetPack)
				admin.PUT("/packs/:id", packHandler.UpdatePack)
				admin.DELETE("/packs/:id", packHandler.DeletePack)

				// Subscription management
				admin.GET("/subscriptions", subscriptionHandler.ListSubscriptions)
				admin.POST("/subscriptions", subscriptionHandler.CreateSubscription)
				admin.GET("/subscriptions/:id", subscriptionHandler.GetSubscription)
				admin.PUT("/subscriptions/:id/approve", subscriptionHandler.ApproveSubscription)
				admin.PUT("/subscriptions/:id/assign", subscriptionHandler.AssignSubscription)
				admin.PUT("/subscriptions/:id/unassign", subscriptionHandler.UnassignSubscription)
				admin.DELETE("/subscriptions/:id", subscriptionHandler.DeleteSubscription)
			}

			// Customer endpoints
			customer := v1.Group("/customer")
			customer.Use(middleware.CustomerOnly())
			{
				customer.GET("/profile", customerHandler.GetProfile)
				customer.PUT("/profile", customerHandler.UpdateProfile)
				customer.GET("/subscription", subscriptionHandler.GetCurrentSubscription)
				customer.POST("/subscription/request", subscriptionHandler.RequestSubscription)
				customer.PUT("/subscription/deactivate", subscriptionHandler.DeactivateSubscription)
				customer.GET("/subscription/history", subscriptionHandler.GetSubscriptionHistory)
			}
		}
	}

	// SDK API routes (API Key authentication)
	sdk := router.Group("/sdk")
	{
		// Public SDK authentication
		sdk.POST("/auth/login", sdkHandler.Login)

		// Protected SDK endpoints (API Key required)
		sdkV1 := sdk.Group("/v1")
		sdkV1.Use(middleware.APIKeyAuth())
		{
			sdkV1.GET("/subscription", sdkHandler.GetCurrentSubscription)
			sdkV1.POST("/subscription/request", sdkHandler.RequestSubscription)
			sdkV1.PUT("/subscription/deactivate", sdkHandler.DeactivateSubscription)
			sdkV1.GET("/subscription/history", sdkHandler.GetSubscriptionHistory)
		}
	}

	return router
}

func createDefaultAdmin(db *database.DB) {
	var count int64
	db.Model(&models.User{}).Where("role = ?", "admin").Count(&count)
	
	if count == 0 {
		admin := &models.User{
			Email:    "admin@example.com",
			Password: "admin123", // In production, this should be hashed
			Role:     "admin",
		}
		
		if err := admin.HashPassword(); err != nil {
			log.Printf("Failed to hash admin password: %v", err)
			return
		}
		
		if err := db.Create(admin).Error; err != nil {
			log.Printf("Failed to create default admin: %v", err)
		} else {
			log.Println("Default admin created: admin@example.com / admin123")
		}
	}
}
