package dtos

// CreateProjectRequest represents the create project payload
type CreateProjectRequest struct {
	Name string `json:"name" binding:"required"`
}

// UpdateProjectRequest represents the update project payload
type UpdateProjectRequest struct {
	Name string `json:"name" binding:"omitempty"`
}
