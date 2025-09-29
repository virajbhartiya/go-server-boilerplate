package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	Server ServerConfig

	// Database configuration
	Database DatabaseConfig

	// API configuration
	API APIConfig

	// Authentication configuration
	Auth AuthConfig

	// Logging configuration
	Logging LoggingConfig

	// Feature flags
	Features FeaturesConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port            string
	Environment     string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	URL                string
	MaxConnections     int
	MaxIdleConnections int
	ConnMaxLifetime    time.Duration
	AutoMigrate        bool
	LogQueries         bool
	PreparedStatements bool
}

// APIConfig holds API-related configuration
type APIConfig struct {
	CorsEnabled    bool
	AllowedOrigins []string
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	JWTSecret           string
	JWTExpiryHours      int
	RefreshTokenEnabled bool
	RefreshTokenExpiry  time.Duration
}

type LoggingConfig struct {
	Level             string
	Format            string
	CallerEnabled     bool
	StacktraceEnabled bool
}

// FeaturesConfig holds feature flags
type FeaturesConfig struct {
	Tracing        bool
	BackgroundJobs bool
}

// LoadConfig loads configuration with defaults and environment overrides
func LoadConfig(env string) (*Config, error) {
	config := getDefaultConfig(env)
	overrideWithEnv(config)

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return config, nil
}

// getDefaultConfig returns a configuration with default values
func getDefaultConfig(env string) *Config {
	config := &Config{
		Server: ServerConfig{
			Port:            "8080",
			Environment:     env,
			ShutdownTimeout: 30 * time.Second,
			ReadTimeout:     15 * time.Second,
			WriteTimeout:    15 * time.Second,
			IdleTimeout:     60 * time.Second,
		},
		Database: DatabaseConfig{
			URL:                "postgres://postgres:postgres@localhost:5432/go_server?sslmode=disable",
			MaxConnections:     25,
			MaxIdleConnections: 5,
			ConnMaxLifetime:    5 * time.Minute,
			AutoMigrate:        true,
			LogQueries:         false,
			PreparedStatements: true,
		},
		API: APIConfig{
			CorsEnabled:    true,
			AllowedOrigins: []string{"*"},
		},
		Auth: AuthConfig{
			JWTSecret:           "your-secret-key-change-in-production",
			JWTExpiryHours:      24,
			RefreshTokenEnabled: true,
			RefreshTokenExpiry:  7 * 24 * time.Hour,
		},
		Logging: LoggingConfig{
			Level:             "info",
			Format:            "json",
			CallerEnabled:     true,
			StacktraceEnabled: false,
		},
		Features: FeaturesConfig{
			Tracing:        false,
			BackgroundJobs: true,
		},
	}

	// Override defaults based on environment
	switch env {
	case "development":
		config.Logging.Level = "debug"
		config.Logging.Format = "console"
		config.Database.LogQueries = true
		config.API.CorsEnabled = true
		config.API.AllowedOrigins = []string{"*"}
	case "test":
		config.Database.URL = "postgres://postgres:postgres@localhost:5432/go_server_test?sslmode=disable"
		config.Logging.Level = "error"
		config.Logging.Format = "console"
		config.Database.LogQueries = false
	case "production":
		config.Logging.Level = "warn"
		config.Logging.Format = "json"
		config.Database.LogQueries = false
		config.API.CorsEnabled = false
		config.API.AllowedOrigins = []string{}
	}

	return config
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	if config.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	if config.Database.URL == "" {
		return fmt.Errorf("database URL is required")
	}

	if config.Auth.JWTSecret == "" || config.Auth.JWTSecret == "your-secret-key-change-in-production" {
		return fmt.Errorf("JWT secret must be set and different from default")
	}

	if config.Auth.JWTExpiryHours <= 0 {
		return fmt.Errorf("JWT expiry hours must be positive")
	}

	// rate limit removed

	return nil
}

// overrideWithEnv overrides configuration values with environment variables
func overrideWithEnv(config *Config) {
	setEnvString := func(envKey string, field *string) {
		if value, exists := os.LookupEnv(envKey); exists {
			*field = value
		}
	}
	setEnvInt := func(envKey string, field *int) {
		if value, exists := os.LookupEnv(envKey); exists {
			if intValue, err := strconv.Atoi(value); err == nil {
				*field = intValue
			}
		}
	}
	setEnvBool := func(envKey string, field *bool) {
		if value, exists := os.LookupEnv(envKey); exists {
			if boolValue, err := strconv.ParseBool(value); err == nil {
				*field = boolValue
			}
		}
	}
	setEnvDuration := func(envKey string, field *time.Duration) {
		if value, exists := os.LookupEnv(envKey); exists {
			if duration, err := time.ParseDuration(value); err == nil {
				*field = duration
			}
		}
	}
	setEnvStringSlice := func(envKey string, field *[]string) {
		if value, exists := os.LookupEnv(envKey); exists && value != "" {
			*field = strings.Split(value, ",")
		}
	}

	// Server configuration
	setEnvString("PORT", &config.Server.Port)
	setEnvString("ENVIRONMENT", &config.Server.Environment)
	setEnvDuration("SHUTDOWN_TIMEOUT", &config.Server.ShutdownTimeout)
	setEnvDuration("READ_TIMEOUT", &config.Server.ReadTimeout)
	setEnvDuration("WRITE_TIMEOUT", &config.Server.WriteTimeout)
	setEnvDuration("IDLE_TIMEOUT", &config.Server.IdleTimeout)
	// SSL removed

	// Database configuration
	setEnvString("DATABASE_URL", &config.Database.URL)
	setEnvInt("DB_MAX_CONNECTIONS", &config.Database.MaxConnections)
	setEnvInt("DB_MAX_IDLE_CONNECTIONS", &config.Database.MaxIdleConnections)
	setEnvDuration("DB_CONN_MAX_LIFETIME", &config.Database.ConnMaxLifetime)
	setEnvBool("DB_AUTO_MIGRATE", &config.Database.AutoMigrate)
	setEnvBool("DB_LOG_QUERIES", &config.Database.LogQueries)
	setEnvBool("DB_PREPARED_STATEMENTS", &config.Database.PreparedStatements)

	// API configuration
	setEnvBool("CORS_ENABLED", &config.API.CorsEnabled)
	setEnvStringSlice("ALLOWED_ORIGINS", &config.API.AllowedOrigins)
	// Rate limiter removed

	// Auth configuration
	setEnvString("JWT_SECRET", &config.Auth.JWTSecret)
	setEnvInt("JWT_EXPIRY_HOURS", &config.Auth.JWTExpiryHours)
	setEnvBool("REFRESH_TOKEN_ENABLED", &config.Auth.RefreshTokenEnabled)
	setEnvDuration("REFRESH_TOKEN_EXPIRY", &config.Auth.RefreshTokenExpiry)

	// Logging configuration
	setEnvString("LOG_LEVEL", &config.Logging.Level)
	setEnvString("LOG_FORMAT", &config.Logging.Format)
	setEnvBool("LOG_CALLER_ENABLED", &config.Logging.CallerEnabled)
	setEnvBool("LOG_STACKTRACE_ENABLED", &config.Logging.StacktraceEnabled)

	// Cache removed

	// Features configuration
	setEnvBool("ENABLE_TRACING", &config.Features.Tracing)
	setEnvBool("ENABLE_BACKGROUND_JOBS", &config.Features.BackgroundJobs)

	// Redis removed
}
