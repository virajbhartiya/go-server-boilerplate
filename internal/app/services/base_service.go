package services

import (
	"context"

	"go-server-boilerplate/internal/app/domain"
	"go-server-boilerplate/internal/app/ports"
)

// BaseService is a generic implementation of the Service interface
type BaseService[T domain.Entity] struct {
	repository ports.Repository[T]
}

// NewBaseService creates a new base service
func NewBaseService[T domain.Entity](repository ports.Repository[T]) *BaseService[T] {
	return &BaseService[T]{
		repository: repository,
	}
}

// Create creates a new entity
func (s *BaseService[T]) Create(ctx context.Context, entity T) error {
	return s.repository.Create(ctx, entity)
}

// GetByID retrieves an entity by its ID
func (s *BaseService[T]) GetByID(ctx context.Context, id uint) (T, error) {
	return s.repository.FindByID(ctx, id)
}

// Update updates an existing entity
func (s *BaseService[T]) Update(ctx context.Context, entity T) error {
	return s.repository.Update(ctx, entity)
}

// Delete removes an entity
func (s *BaseService[T]) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}

// List retrieves entities with pagination
func (s *BaseService[T]) List(ctx context.Context, page, pageSize int) ([]T, int64, error) {
	return s.repository.List(ctx, page, pageSize)
}
