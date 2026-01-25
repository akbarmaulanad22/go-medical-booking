package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go-template-clean-architecture/config"
	deliveryHttp "go-template-clean-architecture/internal/delivery/http"
	"go-template-clean-architecture/internal/delivery/http/handler"
	"go-template-clean-architecture/internal/delivery/http/middleware"
	"go-template-clean-architecture/internal/infrastructure/cache"
	"go-template-clean-architecture/internal/infrastructure/database"
	"go-template-clean-architecture/internal/repository"
	"go-template-clean-architecture/internal/usecase"
	"go-template-clean-architecture/pkg/jwt"
	"go-template-clean-architecture/pkg/validator"

	"github.com/sirupsen/logrus"
)

func main() {
	// Setup logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logrus.Fatalf("Failed to load config: %v", err)
	}

	logrus.Info("Configuration loaded successfully")

	// Initialize database
	db, err := database.NewPostgresConnection(cfg.DB)
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	logrus.Info("Database connected successfully")

	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		logrus.Fatalf("Failed to connect to Redis: %v", err)
	}
	logrus.Info("Redis connected successfully")

	// Initialize JWT service
	jwtService := jwt.NewJWTService(cfg.JWT)

	// Initialize validator
	customValidator := validator.NewValidator()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	productRepo := repository.NewProductRepository(db)

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, jwtService, redisClient)
	productUsecase := usecase.NewProductUsecase(productRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUsecase, customValidator, jwtService)
	productHandler := handler.NewProductHandler(productUsecase, customValidator)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, redisClient)

	// Initialize router
	router := deliveryHttp.NewRouter(authHandler, productHandler, authMiddleware)
	httpRouter := router.Setup()

	// Start server
	serverAddr := fmt.Sprintf(":%s", cfg.App.Port)
	server := &http.Server{
		Addr:    serverAddr,
		Handler: httpRouter,
	}

	// Graceful shutdown
	go func() {
		logrus.Infof("Server starting on port %s", cfg.App.Port)
		logrus.Infof("Environment: %s", cfg.App.Env)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Close database connection
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	// Close Redis connection
	redisClient.Close()

	logrus.Info("Server shutdown complete")
}
