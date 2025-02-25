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

	"go-server/internal/config"
	"go-server/internal/database"
	"go-server/internal/handlers"
	"go-server/internal/logger"
	"go-server/internal/middleware"
	"go-server/internal/repository"
	"go-server/internal/service"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(cfg.LogLevel, cfg.LogJSON)
	defer logger.Sync()

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create a context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database connection
	dbPool, err := pgxpool.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Fatal("Unable to connect to database", zap.Error(err))
	}
	defer dbPool.Close()

	// Run database migrations
	if err := database.RunMigrations(cfg.DatabaseURL); err != nil {
		logger.Fatal("Failed to run database migrations", zap.Error(err))
	}

	// Initialize repository, service, and handler layers
	repo := repository.NewRepository(dbPool)
	svc := service.NewService(repo)
	handler := handlers.NewHandler(svc)

	// Create a WaitGroup for tracking in-flight requests
	var wg sync.WaitGroup

	// Initialize router
	router := gin.New()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Metrics())
	router.Use(middleware.RateLimiter(cfg.RateLimitRequests, cfg.RateLimitDuration))
	router.Use(middleware.Cors(cfg.AllowedOrigins))

	// Add middleware to track in-flight requests
	router.Use(func(c *gin.Context) {
		wg.Add(1)
		defer wg.Done()
		c.Next()
	})

	// Setup routes
	handler.SetupRoutes(router)

	// Add metrics endpoint
	if cfg.MetricsEnabled {
		router.GET(cfg.MetricsPath, gin.WrapH(promhttp.Handler()))
	}

	// Add pprof endpoints in development
	if cfg.Environment == "development" {
		pprof.Register(router)
	}

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Create server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Start server
	go func() {
		logger.Info("Server is starting", zap.String("port", cfg.Port))
		var err error
		if cfg.SSLEnabled {
			err = srv.ListenAndServeTLS(cfg.SSLCertFile, cfg.SSLKeyFile)
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
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

	logger.Info("Server exited")
}