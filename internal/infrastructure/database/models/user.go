package models

import (
	"time"

	"go-server-boilerplate/internal/app/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	domain.BaseEntity
	Email        string     `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string     `gorm:"type:varchar(255);not null" json:"-"`
	FirstName    string     `gorm:"type:varchar(255)" json:"first_name"`
	LastName     string     `gorm:"type:varchar(255)" json:"last_name"`
	Role         string     `gorm:"type:varchar(50);default:'user'" json:"role"`
	LastLogin    *time.Time `json:"last_login"`
	Active       bool       `gorm:"default:true" json:"active"`
}

// TableName overrides the table name
func (User) TableName() string {
	return "users"
}

// BeforeCreate hook is called before creating a user
func (u *User) BeforeCreate(tx *gorm.DB) error {
	// If password hash is not set, return error (password is required)
	if u.PasswordHash == "" {
		return gorm.ErrInvalidData
	}
	return nil
}

// SetPassword hashes the password and sets it to the user
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword checks if the provided password matches the stored hash
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// UpdateLastLogin updates the last login time to the current time
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLogin = &now
}

// DisplayName returns the full name of the user or their email if not available
func (u *User) DisplayName() string {
	if u.FirstName != "" || u.LastName != "" {
		return u.FirstName + " " + u.LastName
	}
	return u.Email
}
