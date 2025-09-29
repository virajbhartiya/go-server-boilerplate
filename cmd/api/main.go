package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"go-server-boilerplate/internal/app/services"
	"go-server-boilerplate/internal/config"
	"go-server-boilerplate/internal/infrastructure/auth"
	"go-server-boilerplate/internal/infrastructure/database"
	"go-server-boilerplate/internal/infrastructure/database/models"
	"go-server-boilerplate/internal/infrastructure/jobs"

	"go-server-boilerplate/internal/interfaces/api"
	"go-server-boilerplate/internal/pkg/logger"
	"go-server-boilerplate/internal/pkg/middleware"
)

// @title Go Server Boilerplate API
// @version 1.0
// @description Go Server Boilerplate API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.example.com/support
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1
// @schemes http https

func main() {
	_ = godotenv.Load()

	// Determine environment
	env := os.Getenv("ENVIRONMENT")
	if env == "" {
		env = "development"
	}

	// Load configuration
	cfg, err := config.LoadConfig(env)
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger.Init(cfg.Logging.Level, cfg.Logging.Format == "json")
	defer logger.Sync()

	logger.Info("Starting application",
		zap.String("environment", env),
		zap.String("port", cfg.Server.Port),
	)

	// No Gin mode; using Gorilla Mux with net/http

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection
	dbConfig := database.Config{
		URL:                cfg.Database.URL,
		MaxConnections:     cfg.Database.MaxConnections,
		MaxIdleConns:       cfg.Database.MaxIdleConnections,
		ConnMaxLifetime:    cfg.Database.ConnMaxLifetime,
		AutoMigrate:        cfg.Database.AutoMigrate,
		LogQueries:         cfg.Database.LogQueries,
		PreparedStatements: cfg.Database.PreparedStatements,
	}

	db, err := database.Connect(dbConfig)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Use ctx in a database ping to silence unused variable
	sqlDB, err := db.DB()
	if err == nil {
		_ = sqlDB.PingContext(ctx)
	}

	// Initialize JWT manager
	jwtManager := auth.NewJWTManager(cfg.Auth.JWTSecret, cfg.Auth.JWTExpiryHours)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtManager)

	// Initialize repositories
	userRepo := database.NewGormRepository[models.User](db)

	// Initialize services
	userService := services.NewBaseService[models.User](userRepo)

	// Initialize handlers
	userHandler := api.NewUserHandler(userService)
	authHandler := api.NewAuthHandler(userService, jwtManager)

	// Initialize background job system if enabled
	var jobDispatcher *jobs.Dispatcher
	if cfg.Features.BackgroundJobs {
		jobDispatcher = jobs.NewDispatcher(5) // 5 workers
		jobDispatcher.Start()
		defer func() {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			jobDispatcher.Stop(shutdownCtx)
		}()

		// Example of adding a scheduled job
		// cleanupJob := jobs.NewScheduledJob("database-cleanup", 24*time.Hour, func(ctx context.Context) error {
		//	 // Perform cleanup tasks
		//	 return nil
		// }, jobDispatcher)
		// cleanupJob.Start()
	}

	// Create a WaitGroup for tracking in-flight requests
	var wg sync.WaitGroup

	// Initialize router
	router := mux.NewRouter()

	// Build middleware chain for net/http
	var handler http.Handler = router

	// Apply middleware in reverse order (innermost first)
	if cfg.API.CorsEnabled {
		handler = middleware.Cors(handler, cfg.API.AllowedOrigins)
	}
	handler = middleware.RecoveryMiddleware(handler)
	handler = middleware.RequestIDMiddleware(handler)

	// Track in-flight requests (outermost)
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		wg.Add(1)
		defer wg.Done()
		handler.ServeHTTP(w, r)
	})

	// Setup routes
	setupRoutesMux(router, authMiddleware, userHandler, authHandler)

	// Health route
	api.RegisterHealthRoutesMux(router)

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server
	go func() {
		logger.Info("Server is starting", zap.String("port", cfg.Server.Port))
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutdown signal received")

	// Cancel the context to notify all operations
	cancel()

	// Give outstanding requests time to complete
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	logger.Info("Shutting down server")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Wait for in-flight requests to complete with timeout
	waitChan := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitChan)
	}()

	select {
	case <-waitChan:
		logger.Info("All requests completed")
	case <-shutdownCtx.Done():
		logger.Warn("Timeout waiting for requests to complete")
	}

	// Close database connection
	if sqlDB != nil {
		if err := sqlDB.Close(); err != nil {
			logger.Error("Error closing database connection", zap.Error(err))
		} else {
			logger.Info("Database connection closed")
		}
	}

	logger.Info("Server exited")
}

// setupRoutes configures all the routes for the application
func setupRoutesMux(r *mux.Router, authMiddleware *middleware.AuthMiddleware, userHandler *api.UserHandler, authHandler *api.AuthHandler) {
	// Register auth routes
	authHandler.RegisterAuthRoutes(r)

	// Register user routes
	userHandler.RegisterUserRoutes(r, authMiddleware)
}
