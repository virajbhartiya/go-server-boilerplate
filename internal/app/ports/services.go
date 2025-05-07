package ports

import (
	"context"

	"go-server-boilerplate/internal/app/domain"
)

// Service defines the base service operations
type Service[T domain.Entity] interface {
	// Create creates a new entity
	Create(ctx context.Context, entity T) error

	// GetByID retrieves an entity by its ID
	GetByID(ctx context.Context, id uint) (T, error)

	// Update updates an existing entity
	Update(ctx context.Context, entity T) error

	// Delete removes an entity
	Delete(ctx context.Context, id uint) error

	// List retrieves entities with pagination
	List(ctx context.Context, page, pageSize int) ([]T, int64, error)
}
