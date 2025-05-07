package ports

import (
	"context"

	"go-server-boilerplate/internal/app/domain"
)

// Repository defines the base repository operations
type Repository[T domain.Entity] interface {
	// Create creates a new entity
	Create(ctx context.Context, entity T) error

	// FindByID retrieves an entity by its ID
	FindByID(ctx context.Context, id uint) (T, error)

	// Update updates an existing entity
	Update(ctx context.Context, entity T) error

	// Delete removes an entity
	Delete(ctx context.Context, id uint) error

	// List retrieves entities with pagination
	List(ctx context.Context, page, pageSize int) ([]T, int64, error)
}

// TransactionManager defines the interface for database transactions
type TransactionManager interface {
	// WithTransaction executes the given function in a transaction
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
