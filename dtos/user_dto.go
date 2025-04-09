package dtos

import (
	"github.com/atqamz/kogase-backend/models"
	"github.com/google/uuid"
)

type CreateUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

type CreateUserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type GetUsersResponse struct {
	Users []GetUserResponse `json:"users"`
}

type GetUserResponse struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Name   string    `json:"name"`
}

type GetUserResponseDetail struct {
	GetUserResponse
	Projects []models.Project `json:"projects"`
}

type UpdateUserRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Name     string `json:"name" binding:"omitempty"`
	Password string `json:"password" binding:"omitempty"`
}

type UpdateUserResponse struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type DeleteUserResponse struct {
	Message string `json:"message"`
}
