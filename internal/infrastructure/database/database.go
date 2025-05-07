package database

import (
	"context"
	"fmt"
	"time"

	"go-server-boilerplate/internal/infrastructure/database/models"
	"go-server-boilerplate/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// Config represents the database configuration
type Config struct {
	URL                string
	MaxConnections     int
	MaxIdleConns       int
	ConnMaxLifetime    time.Duration
	AutoMigrate        bool
	LogQueries         bool
	PreparedStatements bool
}

// Connect establishes a connection to the database
func Connect(cfg Config) (*gorm.DB, error) {
	logger.Info("Connecting to database", zap.String("url", cfg.URL))

	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: false,
		},
		PrepareStmt: cfg.PreparedStatements,
	}

	// Set log level based on configuration
	if cfg.LogQueries {
		gormConfig.Logger = NewGormLogger()
	}

	// Connect to the database
	db, err := gorm.Open(postgres.Open(cfg.URL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(cfg.MaxConnections)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	// Ping the database to verify connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Successfully connected to database")

	// Auto migrate if enabled
	if cfg.AutoMigrate {
		logger.Info("Running auto migrations")
		if err := autoMigrate(db); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}
	}

	return db, nil
}

// autoMigrate automatically migrates the database schema
func autoMigrate(db *gorm.DB) error {
	// Add all models to migrate here
	return db.AutoMigrate(
		&models.User{},
		// Add more models here as needed
	)
}

// Close closes the database connection
func Close(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		logger.Error("Failed to get database connection for closing", zap.Error(err))
		return
	}

	if err := sqlDB.Close(); err != nil {
		logger.Error("Failed to close database connection", zap.Error(err))
		return
	}

	logger.Info("Database connection closed")
}
