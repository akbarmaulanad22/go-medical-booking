package bootstrap

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// App holds all dependencies for the application
type App struct {
	Config      *config.Config
	DB          *gorm.DB
	RedisClient *redis.Client
	Server      *http.Server
}

// New creates a new App instance with all dependencies initialized
func New() (*App, error) {
	app := &App{}

	// Setup logger
	setupLogger()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	app.Config = cfg
	logrus.Info("Configuration loaded successfully")

	// Initialize database
	db, err := database.NewPostgresConnection(cfg.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}
	app.DB = db
	logrus.Info("Database connected successfully")

	// Initialize Redis
	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	app.RedisClient = redisClient
	logrus.Info("Redis connected successfully")

	// Initialize all layers
	server := initializeServer(cfg, db, redisClient)
	app.Server = server

	return app, nil
}

// setupLogger configures the logrus logger
func setupLogger() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)
}

// initializeServer creates and configures the HTTP server
func initializeServer(cfg *config.Config, db *gorm.DB, redisClient *redis.Client) *http.Server {
	// Initialize JWT service
	jwtService := jwt.NewJWTService(cfg.JWT)

	// Initialize validator
	customValidator := validator.NewValidator()

	// Initialize repositories
	userRepo := repository.NewUserRepository()
	roleRepo := repository.NewRoleRepository()
	doctorProfileRepo := repository.NewDoctorProfileRepository()
	patientProfileRepo := repository.NewPatientProfileRepository()
	doctorScheduleRepo := repository.NewDoctorScheduleRepository()

	// Initialize logger
	log := logrus.StandardLogger()

	// Initialize usecases
	authUsecase := usecase.NewAuthUsecase(db, log, userRepo, roleRepo, doctorProfileRepo, patientProfileRepo, jwtService, redisClient)
	doctorProfileUsecase := usecase.NewDoctorProfileUsecase(db, log, userRepo, doctorProfileRepo)
	doctorScheduleUsecase := usecase.NewDoctorScheduleUsecase(db, log, doctorScheduleRepo)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUsecase, customValidator, jwtService)
	doctorHandler := handler.NewDoctorHandler(doctorProfileUsecase, customValidator)
	doctorScheduleHandler := handler.NewDoctorScheduleHandler(doctorScheduleUsecase, customValidator)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, redisClient)
	corsMiddleware := middleware.NewCORSMiddleware()

	// Initialize router
	router := deliveryHttp.NewRouter(authHandler, doctorHandler, doctorScheduleHandler, authMiddleware, corsMiddleware)
	httpRouter := router.Setup()

	// Create server
	serverAddr := fmt.Sprintf(":%s", cfg.App.Port)
	return &http.Server{
		Addr:    serverAddr,
		Handler: httpRouter,
	}
}

// Run starts the HTTP server and handles graceful shutdown
func (app *App) Run() {
	// Start server in goroutine
	go func() {
		logrus.Infof("Server starting on port %s", app.Config.App.Port)
		logrus.Infof("Environment: %s", app.Config.App.Env)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	app.waitForShutdown()
}

// waitForShutdown blocks until an interrupt signal is received
func (app *App) waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server gracefully
	if err := app.Server.Shutdown(ctx); err != nil {
		logrus.Errorf("Server forced to shutdown: %v", err)
	}

	// Close connections
	app.Close()

	logrus.Info("Server shutdown complete")
}

// Close closes all connections (database, redis, etc.)
func (app *App) Close() {
	// Close database connection
	if app.DB != nil {
		sqlDB, err := app.DB.DB()
		if err == nil {
			sqlDB.Close()
		}
	}

	// Close Redis connection
	if app.RedisClient != nil {
		app.RedisClient.Close()
	}
}
