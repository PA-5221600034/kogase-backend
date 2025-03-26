package dtos

import (
	"github.com/atqamz/kogase-backend/models"
	"github.com/google/uuid"
)

// CreateProjectRequest represents the create project payload
type CreateProjectRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateProjectResponse represents the create project response
type CreateProjectResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	ApiKey  string    `json:"api_key"`
	OwnerID uuid.UUID `json:"owner_id"`
}

// GetProjectResponseDetail represents the get project response detail (with Devices and Events)
type GetProjectResponseDetail struct {
	ID      uuid.UUID       `json:"id"`
	Name    string          `json:"name"`
	ApiKey  string          `json:"api_key"`
	OwnerID uuid.UUID       `json:"owner_id"`
	Devices []models.Device `json:"devices"`
	Events  []models.Event  `json:"events"`
}

// GetProjectResponse represents the get project response
type GetProjectResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	ApiKey  string    `json:"api_key"`
	OwnerID uuid.UUID `json:"owner_id"`
}

// GetProjectsResponse represents the get projects response
type GetProjectsResponse struct {
	Projects []GetProjectResponse `json:"projects"`
}

// UpdateProjectRequest represents the update project payload
type UpdateProjectRequest struct {
	Name string `json:"name" binding:"omitempty"`
}

// UpdateProjectResponse represents the update project response
type UpdateProjectResponse struct {
	ID      uuid.UUID `json:"id"`
	Name    string    `json:"name"`
	ApiKey  string    `json:"api_key"`
	OwnerID uuid.UUID `json:"owner_id"`
}

// DeleteProjectResponse represents the delete project response
type DeleteProjectResponse struct {
	Message string `json:"message"`
}
