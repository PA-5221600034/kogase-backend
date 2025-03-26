package dtos

import (
	"time"

	"github.com/google/uuid"
)

// LoginRequest represents the request to login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// MeResponse represents the current user information response
type MeResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LogoutResponse represents the logout response
type LogoutResponse struct {
	Message string `json:"message"`
}
