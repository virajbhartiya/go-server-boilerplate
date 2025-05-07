package testing

import (
	"context"

	"go-server-boilerplate/internal/app/domain"
	"go-server-boilerplate/internal/app/ports"
)

// MockRepository is a generic mock repository for testing
type MockRepository[T domain.Entity] struct {
	CreateFunc   func(ctx context.Context, entity T) error
	FindByIDFunc func(ctx context.Context, id uint) (T, error)
	UpdateFunc   func(ctx context.Context, entity T) error
	DeleteFunc   func(ctx context.Context, id uint) error
	ListFunc     func(ctx context.Context, page, pageSize int) ([]T, int64, error)
}

// Ensure MockRepository implements Repository interface
var _ ports.Repository[domain.Entity] = &MockRepository[domain.Entity]{}

// Create implements Repository.Create
func (m *MockRepository[T]) Create(ctx context.Context, entity T) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return nil
}

// FindByID implements Repository.FindByID
func (m *MockRepository[T]) FindByID(ctx context.Context, id uint) (T, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	var zero T
	return zero, nil
}

// Update implements Repository.Update
func (m *MockRepository[T]) Update(ctx context.Context, entity T) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return nil
}

// Delete implements Repository.Delete
func (m *MockRepository[T]) Delete(ctx context.Context, id uint) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// List implements Repository.List
func (m *MockRepository[T]) List(ctx context.Context, page, pageSize int) ([]T, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, page, pageSize)
	}
	return []T{}, 0, nil
}

// MockService is a generic mock service for testing
type MockService[T domain.Entity] struct {
	CreateFunc  func(ctx context.Context, entity T) error
	GetByIDFunc func(ctx context.Context, id uint) (T, error)
	UpdateFunc  func(ctx context.Context, entity T) error
	DeleteFunc  func(ctx context.Context, id uint) error
	ListFunc    func(ctx context.Context, page, pageSize int) ([]T, int64, error)
}

// Ensure MockService implements Service interface
var _ ports.Service[domain.Entity] = &MockService[domain.Entity]{}

// Create implements Service.Create
func (m *MockService[T]) Create(ctx context.Context, entity T) error {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, entity)
	}
	return nil
}

// GetByID implements Service.GetByID
func (m *MockService[T]) GetByID(ctx context.Context, id uint) (T, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	var zero T
	return zero, nil
}

// Update implements Service.Update
func (m *MockService[T]) Update(ctx context.Context, entity T) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, entity)
	}
	return nil
}

// Delete implements Service.Delete
func (m *MockService[T]) Delete(ctx context.Context, id uint) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

// List implements Service.List
func (m *MockService[T]) List(ctx context.Context, page, pageSize int) ([]T, int64, error) {
	if m.ListFunc != nil {
		return m.ListFunc(ctx, page, pageSize)
	}
	return []T{}, 0, nil
}
