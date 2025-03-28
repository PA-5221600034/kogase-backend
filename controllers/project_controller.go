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
				Error: "User not found",
			}
			c.JSON(http.StatusUnauthorized, response)
			return
		}
		userID = defaultAdmin.ID
	}

	// Bind request body
	var request dtos.CreateProjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Create project
	project := models.Project{
		Name:    request.Name,
		ApiKey:  uuid.New().String(),
		OwnerID: userID.(uuid.UUID),
	}
	if err := pc.DB.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	// Create response DTO
	response := dtos.CreateProjectResponse{
		ID:      project.ID,
		Name:    project.Name,
		ApiKey:  project.ApiKey,
		OwnerID: project.OwnerID,
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get project ID from URL
	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Check if user has access to project (only owner has access)
	if project.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Create response DTO
	response := dtos.GetProjectResponseDetail{
		ID:      project.ID,
		Name:    project.Name,
		ApiKey:  project.ApiKey,
		OwnerID: project.OwnerID,
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get projects, users can only see their own projects
	var projects []models.Project
	if err := pc.DB.Where("owner_id = ?", userID).Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve projects"})
		return
	}

	// Create response DTO
	response := dtos.GetProjectsResponse{
		Projects: make([]dtos.GetProjectResponse, len(projects)),
	}
	for i, project := range projects {
		response.Projects[i] = dtos.GetProjectResponse{
			ID:     project.ID,
			Name:   project.Name,
			ApiKey: project.ApiKey,
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get project ID from URL
	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Check if user has access to project (only owner has access)
	if project.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Bind request body
	var updateReq dtos.UpdateProjectRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Update project
	if updateReq.Name != "" {
		project.Name = updateReq.Name
	}

	if err := pc.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	// Create response DTO
	response := dtos.UpdateProjectResponse{
		ID:      project.ID,
		Name:    project.Name,
		ApiKey:  project.ApiKey,
		OwnerID: project.OwnerID,
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get project ID from URL
	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Check if user has access to project (only owner has access)
	if project.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete project
	if err := pc.DB.Delete(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	// Create response DTO
	response := dtos.DeleteProjectResponse{
		Message: "Project deleted successfully",
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// GetProjectByApiKey returns the project through API key
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
func (pc *ProjectController) GetProjectByApiKey(c *gin.Context) {
	// Get project ID from context (set by ApiKeyMiddleware)
	projectID, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Create response DTO
	response := dtos.GetProjectResponseDetail{
		ID:      project.ID,
		Name:    project.Name,
		ApiKey:  project.ApiKey,
		OwnerID: project.OwnerID,
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
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	// Get project ID from URL
	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project
	var project models.Project
	if err := pc.DB.First(&project, "id = ?", projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Check if user has access to project (only owner has access)
	if project.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Generate new API key
	project.ApiKey = uuid.New().String()

	// Save project
	if err := pc.DB.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate API key"})
		return
	}

	// Create response DTO
	response := dtos.GetProjectResponse{
		ID:      project.ID,
		Name:    project.Name,
		ApiKey:  project.ApiKey,
		OwnerID: project.OwnerID,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}
