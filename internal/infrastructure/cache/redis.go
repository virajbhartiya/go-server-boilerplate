package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go-server-boilerplate/internal/config"

	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

// RedisClient implements the Cache interface using Redis
type RedisClient struct {
	client *redis.Client
	log    *zap.Logger
}

// RedisCache implements the Cache interface using Redis
type RedisCache struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client instance
func NewRedisClient(cfg *config.Config, log *zap.Logger) (Cache, error) {
	// Use default Redis settings
	host := "localhost"
	port := "6379"
	password := ""
	db := 0

	// Use Redis URL from config if enabled
	if cfg.Cache.Enabled && cfg.Cache.RedisURL != "" {
		// Simple parsing - in production code, use a proper URL parser
		// Expected format: redis://user:password@host:port/db
		url := cfg.Cache.RedisURL
		if strings.HasPrefix(url, "redis://") {
			// Extract connection details from URL
			// This is simplified; real code should use net/url package
			hostPort := url
			if strings.Contains(url, "@") {
				parts := strings.Split(url, "@")
				hostPort = parts[1]
				if strings.Contains(parts[0], ":") {
					password = strings.Split(parts[0], ":")[1]
				}
			}

			if strings.Contains(hostPort, ":") {
				hostPortParts := strings.Split(hostPort, ":")
				host = hostPortParts[0]
				port = strings.Split(hostPortParts[1], "/")[0]
			}
		}
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		log.Error("Failed to connect to Redis", zap.Error(err))
		return nil, err
	}

	log.Info("Successfully connected to Redis")
	return &RedisClient{
		client: redisClient,
		log:    log,
	}, nil
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(cfg *config.Config) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// Test connection
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: client,
	}, nil
}

// Get retrieves a value from Redis by key
func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Key does not exist
		}
		return "", err
	}
	return result, nil
}

// Set stores a value in Redis with the specified expiration
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	default:
		// Marshal non-string values to JSON
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		strValue = string(jsonBytes)
	}

	return r.client.Set(ctx, key, strValue, expiration).Err()
}

// Delete removes a value from Redis by key
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return result > 0, nil
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// FlushDB flushes the entire Redis database - use with caution!
func (r *RedisClient) FlushDB(ctx context.Context) error {
	_, err := r.client.FlushDB(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to flush Redis DB: %w", err)
	}
	return nil
}

// GetClient returns the underlying Redis client
func (r *RedisClient) GetClient() *redis.Client {
	return r.client
}

// Get retrieves a value from the cache by key
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key %s not found", key)
	} else if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}

	return val, nil
}

// Set stores a value in the cache with an optional expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var strValue string

	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		// JSON serialize other types
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		strValue = string(bytes)
	}

	err := r.client.Set(ctx, key, strValue, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	return nil
}

// Delete removes a value from the cache by key
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	return nil
}

// Exists checks if a key exists in the cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	val, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}

	return val > 0, nil
}

// Close closes the Redis connection
func (r *RedisCache) Close() error {
	return r.client.Close()
}
