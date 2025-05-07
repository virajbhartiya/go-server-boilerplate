package cache

import (
	"context"
	"time"
)

// Cache defines the interface for interacting with a cache store
type Cache interface {
	// Get retrieves a value from the cache by key
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value in the cache with an optional expiration
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error

	// Delete removes a value from the cache by key
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// Close closes the cache connection
	Close() error
}
