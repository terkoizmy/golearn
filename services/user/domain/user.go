package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a user entity in the system
type User struct {
	ID        string    `gorm:"type:uuid;primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Name      string    `gorm:"not null" json:"name"`
	Password  string    `gorm:"not null" json:"-"` // Never expose password in JSON
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return nil
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
}

// RegisterRequest represents user registration input
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Name     string `json:"name" binding:"required,min=2,max=100" example:"John Doe"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// LoginRequest represents user login input
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// LoginResponse represents successful login output
type LoginResponse struct {
	Token string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  *User  `json:"user"`
}

// UserResponse represents user data without sensitive info
type UserResponse struct {
	ID        string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email     string    `json:"email" example:"user@example.com"`
	Name      string    `json:"name" example:"John Doe"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"invalid credentials"`
}
