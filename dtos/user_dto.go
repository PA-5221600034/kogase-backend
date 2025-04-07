package dtos

import (
	"github.com/atqamz/kogase-backend/models"
	"github.com/google/uuid"
)

// CreateUserRequest represents the registration payload
type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

// CreateUserResponse represents the response for creating a user
type CreateUserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// GetUserResponseDetail represents the response for getting a user
type GetUserResponseDetail struct {
	UserID   uuid.UUID        `json:"user_id"`
	Email    string           `json:"email"`
	Name     string           `json:"name"`
	Projects []models.Project `json:"projects"`
}

// GetUserResponse represents the response for getting a user
type GetUserResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
}

// GetUsersResponse represents the response for getting all users
type GetUsersResponse struct {
	Users []GetUserResponse `json:"users"`
}

// UpdateUserRequest represents data for updating user information
type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Name     string `json:"name" binding:"omitempty"`
	Password string `json:"password" binding:"omitempty"`
}

// UpdateUserResponse represents the response for updating a user
type UpdateUserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// DeleteUserResponse represents the response for deleting a user
type DeleteUserResponse struct {
	Message string `json:"message"`
}
