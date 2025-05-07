package database

import (
	"context"
	"errors"

	"go-server-boilerplate/internal/app/domain"
	"go-server-boilerplate/internal/pkg/logger"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// GormRepository is a generic implementation of the Repository interface using GORM
type GormRepository[T domain.Entity] struct {
	db *gorm.DB
}

// NewGormRepository creates a new GORM repository
func NewGormRepository[T domain.Entity](db *gorm.DB) *GormRepository[T] {
	return &GormRepository[T]{
		db: db,
	}
}

// withContext adds context to the GORM DB instance
func (r *GormRepository[T]) withContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

// Create creates a new entity
func (r *GormRepository[T]) Create(ctx context.Context, entity T) error {
	result := r.withContext(ctx).Create(entity)
	if result.Error != nil {
		logger.Error("Failed to create entity", zap.Error(result.Error))
		return result.Error
	}
	return nil
}

// FindByID retrieves an entity by its ID
func (r *GormRepository[T]) FindByID(ctx context.Context, id uint) (T, error) {
	var entity T
	result := r.withContext(ctx).First(&entity, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			var zero T
			return zero, errors.New("entity not found")
		}
		logger.Error("Failed to find entity by ID", zap.Uint("id", id), zap.Error(result.Error))
		return entity, result.Error
	}
	return entity, nil
}

// Update updates an existing entity
func (r *GormRepository[T]) Update(ctx context.Context, entity T) error {
	result := r.withContext(ctx).Save(entity)
	if result.Error != nil {
		logger.Error("Failed to update entity", zap.Error(result.Error))
		return result.Error
	}
	return nil
}

// Delete removes an entity
func (r *GormRepository[T]) Delete(ctx context.Context, id uint) error {
	var entity T
	result := r.withContext(ctx).Delete(&entity, id)
	if result.Error != nil {
		logger.Error("Failed to delete entity", zap.Uint("id", id), zap.Error(result.Error))
		return result.Error
	}
	return nil
}

// List retrieves entities with pagination
func (r *GormRepository[T]) List(ctx context.Context, page, pageSize int) ([]T, int64, error) {
	var entities []T
	var count int64

	// Calculate offset
	offset := (page - 1) * pageSize

	// Get total count
	if err := r.withContext(ctx).Model(new(T)).Count(&count).Error; err != nil {
		logger.Error("Failed to count entities", zap.Error(err))
		return nil, 0, err
	}

	// Get paginated results
	result := r.withContext(ctx).
		Offset(offset).
		Limit(pageSize).
		Find(&entities)

	if result.Error != nil {
		logger.Error("Failed to list entities", zap.Error(result.Error))
		return nil, 0, result.Error
	}

	return entities, count, nil
}

// WithTransaction executes the given function in a transaction
func (r *GormRepository[T]) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create a new context with the transaction
		txCtx := context.WithValue(ctx, txKeyType{}, tx)
		return fn(txCtx)
	})
}

type txKeyType struct{}
