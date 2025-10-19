package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/terkoizmy/golearn/internal/config"
	"github.com/terkoizmy/golearn/internal/database"
	"github.com/terkoizmy/golearn/pkg/pb/user"
	"github.com/terkoizmy/golearn/services/user/domain"
	"github.com/terkoizmy/golearn/services/user/handler"
	"github.com/terkoizmy/golearn/services/user/repository"
	"github.com/terkoizmy/golearn/services/user/service"
	"google.golang.org/grpc"

	_ "github.com/terkoizmy/golearn/docs" // Import generated docs
)

// @title GoLearn User Service API
// @version 1.0
// @description Microservice for user authentication and management
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.golearn.io/support
// @contact.email support@golearn.io

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8081
// @BasePath /
// @schemes http https

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database with GORM
	db, err := database.NewGormDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Auto migrate database schema
	if err := database.AutoMigrate(db, &domain.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Initialize layers
	userRepo := repository.NewUserRepository(db)
	userService := service.NewUserService(userRepo, cfg.JWTSecret)
	httpHandler := handler.NewHTTPHandler(userService)
	grpcHandler := handler.NewGRPCHandler(userService)

	// Setup HTTP server
	router := gin.Default()
	setupRoutes(router, httpHandler)

	httpServer := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: router,
	}

	// Setup gRPC server
	grpcServer := grpc.NewServer()
	user.RegisterUserServiceServer(grpcServer, grpcHandler)

	// Start HTTP server
	go func() {
		log.Printf("🚀 HTTP server starting on port %s", cfg.HTTPPort)
		log.Printf("📚 Swagger docs available at http://localhost:%s/swagger/index.html", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port: %v", err)
		}
		log.Printf("🚀 gRPC server starting on port %s", cfg.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down servers...")

	// Shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Stop gRPC server
	grpcServer.GracefulStop()

	log.Println("✅ Servers stopped gracefully")
}

func setupRoutes(router *gin.Engine, h *handler.HTTPHandler) {
	// Swagger documentation
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "user-service",
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes
		api.POST("/register", h.Register)
		api.POST("/login", h.Login)

		// Protected routes
		api.GET("/users/:id", h.GetUser)
	}
}