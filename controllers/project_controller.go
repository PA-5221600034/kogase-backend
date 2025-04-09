package controllers

import (
	"net/http"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ProjectController handles project-related endpoints
type ProjectController struct {
	DB *gorm.DB
}

// NewProjectController creates a new ProjectController instance
func NewProjectController(db *gorm.DB) *ProjectController {
	return &ProjectController{DB: db}
}

// CreateProject creates a new project
// @Summary Create project
// @Description Create a new project
// @Tags projects
// @Accept json
// @Produce json
// @Param project body dtos.CreateProjectRequest true "Project details"
// @Security BearerAuth
// @Success 201 {object} models.Project
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/projects [post]
func (pc *ProjectController) CreateProject(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		var defaultAdmin models.User
		if err := pc.DB.Where("email = ?", "admin@kogase.io").First(&defaultAdmin).Error; err != nil {
			response := dtos.ErrorResponse{
				Message: "User not found",
			}
			c.JSON(http.StatusUnauthorized, response)
			return
		}
		userID = defaultAdmin.ID
	}

	// Bind request body
	var request dtos.CreateProjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Create project
	project := models.Project{
		Name:    request.Name,
		ApiKey:  uuid.New().String(),
		OwnerID: userID.(uuid.UUID),
	}
	if err := pc.DB.Create(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to create project",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.CreateProjectResponse{
		ProjectID: project.ID,
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID,
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
	}

	// Return response
	c.JSON(http.StatusCreated, response)
}

// GetProject returns a specific project by ID
// @Summary Get project
// @Description Get a specific project by ID
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Security BearerAuth
// @Success 200 {object} models.Project
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/projects/{id} [get]
func (pc *ProjectController) GetProject(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get project ID from URL
	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid project ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Check if user has access to project (only owner has access)
	if project.OwnerID != userID.(uuid.UUID) {
		response := dtos.ErrorResponse{
			Message: "Access denied",
		}
		c.JSON(http.StatusForbidden, response)
		return
	}

	// Create response DTO
	response := dtos.GetProjectResponseDetail{
		ProjectID: project.ID,
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID,
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
		Devices: project.Devices,
		Events:  project.Events,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// GetProjects returns all projects accessible by the current user
// @Summary List projects
// @Description Get all projects accessible by the current user
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Project
// @Failure 401 {object} map[string]string
// @Router /api/v1/projects [get]
func (pc *ProjectController) GetProjects(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var projects []models.Project
	if err := pc.DB.Find(&projects).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to retrieve projects",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.GetProjectsResponse{
		Projects: make([]dtos.GetProjectResponse, len(projects)),
	}
	for i, project := range projects {
		response.Projects[i] = dtos.GetProjectResponse{
			ProjectID: project.ID,
			Name:      project.Name,
			ApiKey:    project.ApiKey,
			Owner: dtos.OwnerDto{
				ID:    project.Owner.ID,
				Email: project.Owner.Email,
				Name:  project.Owner.Name,
			},
		}
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// UpdateProject updates a project
// @Summary Update project
// @Description Update a project's details
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Param project body dtos.UpdateProjectRequest true "Project details"
// @Security BearerAuth
// @Success 200 {object} models.Project
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/projects/{id} [patch]
func (pc *ProjectController) UpdateProject(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get project ID from URL
	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid project ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Bind request body
	var updateReq dtos.UpdateProjectRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update project
	if updateReq.Name != "" {
		project.Name = updateReq.Name
	}

	if err := pc.DB.Save(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to update project",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.UpdateProjectResponse{
		ProjectID: project.ID,
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID,
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// DeleteProject deletes a project
// @Summary Delete project
// @Description Delete a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/projects/{id} [delete]
func (pc *ProjectController) DeleteProject(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get project ID from URL
	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid project ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Delete project
	if err := pc.DB.Delete(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to delete project",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.DeleteProjectResponse{
		Message: "Project deleted successfully",
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// GetProjectWithApiKey returns the project through API key
// @Summary Get project by API key
// @Description Get project by API key
// @Tags projects
// @Accept json
// @Produce json
// @Param api_key path string true "Project API Key"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/projects/apikey [get]
func (pc *ProjectController) GetProjectWithApiKey(c *gin.Context) {
	// Get project ID from context (set by ApiKeyMiddleware)
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Create response DTO
	response := dtos.GetProjectResponseDetail{
		ProjectID: project.ID,
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID,
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
		Devices: project.Devices,
		Events:  project.Events,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// RegenerateApiKey regenerates the API key for a project
// @Summary Regenerate API key
// @Description Regenerate the API key for a project
// @Tags projects
// @Accept json
// @Produce json
// @Param id path string true "Project ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 403 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/projects/{id}/apikey [post]
func (pc *ProjectController) RegenerateApiKey(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get project ID from URL
	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid project ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Generate new API key
	project.ApiKey = uuid.New().String()

	// Save project
	if err := pc.DB.Save(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to regenerate API key",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.GetProjectResponse{
		ProjectID: project.ID,
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID,
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
	}

	// Return response
	c.JSON(http.StatusOK, response)
}
