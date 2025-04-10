package dtos

import (
	"github.com/atqamz/kogase-backend/models"
)

type CreateProjectRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateProjectResponse struct {
	ProjectID string   `json:"project_id"`
	Name      string   `json:"name"`
	ApiKey    string   `json:"api_key"`
	Owner     OwnerDto `json:"owner"`
}

type GetProjectsResponse struct {
	Projects []GetProjectResponse `json:"projects"`
}

type GetProjectResponse struct {
	ProjectID string   `json:"project_id"`
	Name      string   `json:"name"`
	ApiKey    string   `json:"api_key"`
	Owner     OwnerDto `json:"owner"`
}

type GetProjectResponseDetail struct {
	GetProjectResponse
	Devices []models.Device `json:"devices"`
	Events  []models.Event  `json:"events"`
}

type UpdateProjectRequest struct {
	Name string `json:"name" binding:"omitempty"`
}

type UpdateProjectResponse struct {
	ProjectID string   `json:"project_id"`
	Name      string   `json:"name"`
	ApiKey    string   `json:"api_key"`
	Owner     OwnerDto `json:"owner"`
}

type DeleteProjectResponse struct {
	Message string `json:"message"`
}

type OwnerDto struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
