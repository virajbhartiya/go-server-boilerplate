package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	// Server configuration
	Port            string
	Environment     string
	ShutdownTimeout time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration

	// Database configuration
	DatabaseURL          string
	MaxDBConnections     int
	MaxDBIdleConnections int
	DBConnMaxLifetime    time.Duration

	// Security
	JWTSecret      string
	JWTExpiryHours int
	AllowedOrigins []string
	TrustedProxies []string
	SSLEnabled     bool
	SSLCertFile    string
	SSLKeyFile     string

	// Rate limiting
	RateLimitRequests int
	RateLimitDuration time.Duration

	// Logging
	LogLevel string
	LogJSON  bool

	// Monitoring
	MetricsEnabled bool
	MetricsPath    string
}

// Load reads configuration from environment variables
func Load() *Config {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		// Server
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("ENVIRONMENT", "development"),
		ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 15*time.Second),
		ReadTimeout:     getDurationEnv("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:    getDurationEnv("WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:     getDurationEnv("IDLE_TIMEOUT", 60*time.Second),

		// Database
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"),
		MaxDBConnections:     getIntEnv("MAX_DB_CONNECTIONS", 20),
		MaxDBIdleConnections: getIntEnv("MAX_DB_IDLE_CONNECTIONS", 5),
		DBConnMaxLifetime:    getDurationEnv("DB_CONN_MAX_LIFETIME", 5*time.Minute),

		// Security
		JWTSecret:      getEnv("JWT_SECRET", "your-secret-key"),
		JWTExpiryHours: getIntEnv("JWT_EXPIRY_HOURS", 24),
		AllowedOrigins: getSliceEnv("ALLOWED_ORIGINS", []string{"*"}),
		TrustedProxies: getSliceEnv("TRUSTED_PROXIES", []string{"127.0.0.1"}),
		SSLEnabled:     getBoolEnv("SSL_ENABLED", false),
		SSLCertFile:    getEnv("SSL_CERT_FILE", ""),
		SSLKeyFile:     getEnv("SSL_KEY_FILE", ""),

		// Rate limiting
		RateLimitRequests: getIntEnv("RATE_LIMIT_REQUESTS", 100),
		RateLimitDuration: getDurationEnv("RATE_LIMIT_DURATION", time.Minute),

		// Logging
		LogLevel: getEnv("LOG_LEVEL", "info"),
		LogJSON:  getBoolEnv("LOG_JSON", false),

		// Monitoring
		MetricsEnabled: getBoolEnv("METRICS_ENABLED", true),
		MetricsPath:    getEnv("METRICS_PATH", "/metrics"),
	}

	return config
}

// Helper functions to get environment variables with default values
func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value, exists := os.LookupEnv(key); exists {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getSliceEnv(key string, defaultValue []string) []string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return split(value)
	}
	return defaultValue
}

func split(s string) []string {
	var result []string
	current := ""
	inQuotes := false

	for _, char := range s {
		switch char {
		case '"':
			inQuotes = !inQuotes
		case ',':
			if !inQuotes {
				if current != "" {
					result = append(result, current)
					current = ""
				}
			} else {
				current += string(char)
			}
		default:
			current += string(char)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
