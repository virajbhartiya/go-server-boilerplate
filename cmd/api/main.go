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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"

	"go-server-boilerplate/internal/config"
	"go-server-boilerplate/internal/infrastructure/auth"
	"go-server-boilerplate/internal/infrastructure/database"
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

	// Set Gin mode
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

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
		PreparedStatements: cfg.GORM.PreparedStatements,
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
	router := gin.New()

	// Add middleware
	router.Use(middleware.Recovery())
	router.Use(middleware.RequestID())

	// Add API rate limiter if enabled
	if cfg.API.RateLimiterEnabled {
		router.Use(middleware.RateLimiter(cfg.API.RateLimitRequests, cfg.API.RateLimitDuration))
	}

	// Add CORS if enabled
	if cfg.API.CorsEnabled {
		router.Use(middleware.CorsMiddleware(cfg.API.AllowedOrigins))
	}

	// Add middleware to track in-flight requests
	router.Use(func(c *gin.Context) {
		wg.Add(1)
		defer wg.Done()
		c.Next()
	})

	// Add Swagger documentation
	if env != "production" {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Setup routes
	setupRoutes(router, authMiddleware)

	// Add health check endpoint
	api.RegisterHealthRoutes(router)

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server
	go func() {
		logger.Info("Server is starting", zap.String("port", cfg.Server.Port))
		var err error
		if cfg.Server.SSLEnabled {
			err = srv.ListenAndServeTLS(cfg.Server.SSLCertFile, cfg.Server.SSLKeyFile)
		} else {
			err = srv.ListenAndServe()
		}
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
func setupRoutes(r *gin.Engine, authMiddleware *middleware.AuthMiddleware) {
	// Public routes
	v1 := r.Group("/api/v1")
	{
		// Auth routes
		_ = v1.Group("/auth") // Using _ to silence unused variable warning

		// Protected routes
		protected := v1.Group("/")
		protected.Use(authMiddleware.AuthRequired())
		{
			// User routes
			_ = protected.Group("/users") // Using _ to silence unused variable warning

			// Admin routes
			adminGroup := protected.Group("/admin")
			adminGroup.Use(authMiddleware.RoleRequired("admin"))
			_ = adminGroup // Using _ to silence unused variable warning
		}
	}
}
