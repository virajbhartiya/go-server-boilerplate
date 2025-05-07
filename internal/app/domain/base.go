package domain

import (
	"time"
)

// Entity represents a domain entity with basic identity
type Entity interface {
	GetID() uint
}

// BaseEntity provides common fields for all entities
type BaseEntity struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// GetID returns the ID of the entity
func (b BaseEntity) GetID() uint {
	return b.ID
}
