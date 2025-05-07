package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	Server struct {
		Port            string        `toml:"port"`
		Environment     string        `toml:"environment"`
		ShutdownTimeout time.Duration `toml:"shutdown_timeout"`
		ReadTimeout     time.Duration `toml:"read_timeout"`
		WriteTimeout    time.Duration `toml:"write_timeout"`
		IdleTimeout     time.Duration `toml:"idle_timeout"`
		SSLEnabled      bool          `toml:"ssl_enabled"`
		SSLCertFile     string        `toml:"ssl_cert_file"`
		SSLKeyFile      string        `toml:"ssl_key_file"`
	} `toml:"server"`

	// Database configuration
	Database struct {
		URL                string        `toml:"url"`
		MaxConnections     int           `toml:"max_connections"`
		MaxIdleConnections int           `toml:"max_idle_connections"`
		ConnMaxLifetime    time.Duration `toml:"conn_max_lifetime"`
		AutoMigrate        bool          `toml:"auto_migrate"`
		LogQueries         bool          `toml:"log_queries"`
		PreparedStatements bool          `toml:"prepared_statements"`
	} `toml:"database"`

	// GORM configuration
	GORM struct {
		LogLevel               string `toml:"log_level"`
		PreparedStatements     bool   `toml:"prepared_stmt"`
		SkipDefaultTransaction bool   `toml:"skip_default_transaction"`
	} `toml:"gorm"`

	// API configuration
	API struct {
		CorsEnabled        bool          `toml:"cors_enabled"`
		AllowedOrigins     []string      `toml:"allowed_origins"`
		RateLimiterEnabled bool          `toml:"rate_limiter_enabled"`
		RateLimitRequests  int           `toml:"rate_limit_requests"`
		RateLimitDuration  time.Duration `toml:"rate_limit_duration"`
	} `toml:"api"`

	// Authentication configuration
	Auth struct {
		JWTSecret           string        `toml:"jwt_secret"`
		JWTExpiryHours      int           `toml:"jwt_expiry_hours"`
		RefreshTokenEnabled bool          `toml:"refresh_token_enabled"`
		RefreshTokenExpiry  time.Duration `toml:"refresh_token_expiry"`
	} `toml:"auth"`

	// Logging configuration
	Logging struct {
		Level             string `toml:"level"`
		Format            string `toml:"format"`
		CallerEnabled     bool   `toml:"caller_enabled"`
		StacktraceEnabled bool   `toml:"stacktrace_enabled"`
	} `toml:"logging"`

	// Cache configuration
	Cache struct {
		Enabled    bool          `toml:"enabled"`
		RedisURL   string        `toml:"redis_url"`
		DefaultTTL time.Duration `toml:"default_ttl"`
	} `toml:"cache"`

	// Feature flags
	Features struct {
		Tracing        bool `toml:"tracing"`
		BackgroundJobs bool `toml:"background_jobs"`
	} `toml:"features"`

	// Redis configuration
	Redis struct {
		Host     string `toml:"host"`
		Port     string `toml:"port"`
		Password string `toml:"password"`
		DB       int    `toml:"db"`
	} `toml:"redis"`
}

func LoadConfig(env string) (*Config, error) {
	var config Config
	baseConfigPath := filepath.Join("configs", "config.toml")
	if _, err := os.Stat(baseConfigPath); err == nil {
		if _, err := toml.DecodeFile(baseConfigPath, &config); err != nil {
			return nil, fmt.Errorf("failed to load base config: %w", err)
		}
	}
	envConfigPath := filepath.Join("configs", fmt.Sprintf("%s.toml", env))
	if _, err := os.Stat(envConfigPath); err == nil {
		if _, err := toml.DecodeFile(envConfigPath, &config); err != nil {
			return nil, fmt.Errorf("failed to load environment config: %w", err)
		}
	}
	overrideWithEnv(&config)
	return &config, nil
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
	setEnvString("PORT", &config.Server.Port)
	setEnvString("ENVIRONMENT", &config.Server.Environment)
	setEnvDuration("SHUTDOWN_TIMEOUT", &config.Server.ShutdownTimeout)
	setEnvDuration("READ_TIMEOUT", &config.Server.ReadTimeout)
	setEnvDuration("WRITE_TIMEOUT", &config.Server.WriteTimeout)
	setEnvDuration("IDLE_TIMEOUT", &config.Server.IdleTimeout)
	setEnvBool("SSL_ENABLED", &config.Server.SSLEnabled)
	setEnvString("SSL_CERT_FILE", &config.Server.SSLCertFile)
	setEnvString("SSL_KEY_FILE", &config.Server.SSLKeyFile)
	setEnvString("DATABASE_URL", &config.Database.URL)
	setEnvInt("DB_MAX_CONNECTIONS", &config.Database.MaxConnections)
	setEnvInt("DB_MAX_IDLE_CONNECTIONS", &config.Database.MaxIdleConnections)
	setEnvDuration("DB_CONN_MAX_LIFETIME", &config.Database.ConnMaxLifetime)
	setEnvBool("DB_AUTO_MIGRATE", &config.Database.AutoMigrate)
	setEnvBool("DB_LOG_QUERIES", &config.Database.LogQueries)
	setEnvBool("DB_PREPARED_STATEMENTS", &config.Database.PreparedStatements)
	setEnvString("JWT_SECRET", &config.Auth.JWTSecret)
	setEnvInt("JWT_EXPIRY_HOURS", &config.Auth.JWTExpiryHours)
	setEnvBool("REFRESH_TOKEN_ENABLED", &config.Auth.RefreshTokenEnabled)
	setEnvDuration("REFRESH_TOKEN_EXPIRY", &config.Auth.RefreshTokenExpiry)
	setEnvBool("CORS_ENABLED", &config.API.CorsEnabled)
	setEnvStringSlice("ALLOWED_ORIGINS", &config.API.AllowedOrigins)
	setEnvBool("RATE_LIMITER_ENABLED", &config.API.RateLimiterEnabled)
	setEnvInt("RATE_LIMIT_REQUESTS", &config.API.RateLimitRequests)
	setEnvDuration("RATE_LIMIT_DURATION", &config.API.RateLimitDuration)
	setEnvString("LOG_LEVEL", &config.Logging.Level)
	setEnvString("LOG_FORMAT", &config.Logging.Format)
	setEnvBool("ENABLE_CACHE", &config.Cache.Enabled)
	setEnvString("REDIS_URL", &config.Cache.RedisURL)
	setEnvDuration("CACHE_TTL", &config.Cache.DefaultTTL)
	setEnvBool("ENABLE_TRACING", &config.Features.Tracing)
	setEnvBool("ENABLE_BACKGROUND_JOBS", &config.Features.BackgroundJobs)
	setEnvString("REDIS_HOST", &config.Redis.Host)
	setEnvString("REDIS_PORT", &config.Redis.Port)
	setEnvString("REDIS_PASSWORD", &config.Redis.Password)
	setEnvInt("REDIS_DB", &config.Redis.DB)
}
