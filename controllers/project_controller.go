package controllers

import (
	"net/http"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectController struct {
	DB *gorm.DB
}

func NewProjectController(db *gorm.DB) *ProjectController {
	return &ProjectController{DB: db}
}

// CreateProject godoc
// @Summary Create new project
// @Description Create a new telemetry project
// @Tags projects
// @Accept json
// @Produce json
// @Param project body dtos.CreateProjectRequest true "Project details"
// @Success 201 {object} dtos.CreateProjectResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /projects [post]
func (pc *ProjectController) CreateProject(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		var defaultAdmin models.User
		if err := pc.DB.Model(&models.User{}).
			Where("email = ?", "admin@kogase.io").
			First(&defaultAdmin).Error; err != nil {
			response := dtos.ErrorResponse{
				Message: "User not found",
			}
			c.JSON(http.StatusUnauthorized, response)
			return
		}
		userID = defaultAdmin.ID
	}

	var request dtos.CreateProjectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

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

	resultResponse := dtos.CreateProjectResponse{
		ProjectID: project.ID.String(),
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID.String(),
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
	}

	c.JSON(http.StatusCreated, resultResponse)
}

// GetProjects godoc
// @Summary Get all projects
// @Description Retrieve a list of all projects
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.GetProjectsResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /projects [get]
func (pc *ProjectController) GetProjects(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var projects []models.Project
	if err := pc.DB.Model(&models.Project{}).
		Find(&projects).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to retrieve projects",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.GetProjectsResponse{
		Projects: make([]dtos.GetProjectResponse, len(projects)),
	}
	for i, project := range projects {
		resultResponse.Projects[i] = dtos.GetProjectResponse{
			ProjectID: project.ID.String(),
			Name:      project.Name,
			ApiKey:    project.ApiKey,
			Owner: dtos.OwnerDto{
				ID:    project.Owner.ID.String(),
				Email: project.Owner.Email,
				Name:  project.Owner.Name,
			},
		}
	}

	c.JSON(http.StatusOK, resultResponse)
}

// GetProject godoc
// @Summary Get a project by ID
// @Description Retrieve a specific project by its ID
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {object} dtos.GetProjectResponseDetail
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /projects/{id} [get]
func (pc *ProjectController) GetProject(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid project ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var project models.Project
	if err := pc.DB.Model(&models.Project{}).
		Where("id = ?", projectID).
		First(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	resultResponse := dtos.GetProjectResponseDetail{
		GetProjectResponse: dtos.GetProjectResponse{
			ProjectID: project.ID.String(),
			Name:      project.Name,
			ApiKey:    project.ApiKey,
			Owner: dtos.OwnerDto{
				ID:    project.Owner.ID.String(),
				Email: project.Owner.Email,
				Name:  project.Owner.Name,
			},
		},
		Devices: project.Devices,
		Events:  project.Events,
	}

	c.JSON(http.StatusOK, resultResponse)
}

// UpdateProject godoc
// @Summary Update a project
// @Description Update a project's details
// @Tags projects
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Param project body dtos.UpdateProjectRequest true "Updated project details"
// @Success 200 {object} dtos.UpdateProjectResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /projects/{id} [patch]
func (pc *ProjectController) UpdateProject(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid project ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var project models.Project
	if err := pc.DB.Model(&models.Project{}).
		Where("id = ?", projectID).
		First(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	var updateReq dtos.UpdateProjectRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

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

	resultResponse := dtos.UpdateProjectResponse{
		ProjectID: project.ID.String(),
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID.String(),
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
	}

	c.JSON(http.StatusOK, resultResponse)
}

// DeleteProject godoc
// @Summary Delete a project
// @Description Delete a project by its ID
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {object} dtos.DeleteProjectResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /projects/{id} [delete]
func (pc *ProjectController) DeleteProject(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid project ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var project models.Project
	if err := pc.DB.Model(&models.Project{}).
		Where("id = ?", projectID).
		First(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	if err := pc.DB.Delete(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to delete project",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.DeleteProjectResponse{
		Message: "Project deleted successfully",
	}

	c.JSON(http.StatusOK, resultResponse)
}

// RegenerateApiKey godoc
// @Summary Regenerate API key
// @Description Generate a new API key for a project
// @Tags projects
// @Produce json
// @Security BearerAuth
// @Param id path string true "Project ID"
// @Success 200 {object} dtos.GetProjectResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /projects/{id}/apikey [post]
func (pc *ProjectController) RegenerateApiKey(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	id := c.Param("id")
	projectID, err := uuid.Parse(id)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid project ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var project models.Project
	if err := pc.DB.Model(&models.Project{}).
		Where("id = ?", projectID).
		First(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	project.ApiKey = uuid.New().String()

	if err := pc.DB.Save(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to regenerate API key",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.GetProjectResponse{
		ProjectID: project.ID.String(),
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID.String(),
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
	}

	c.JSON(http.StatusOK, resultResponse)
}

// GetProjectWithApiKey godoc
// @Summary Get project with API key
// @Description Get project details using an API key for authentication
// @Tags projects
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dtos.GetProjectResponseDetail
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /projects/apikey [get]
func (pc *ProjectController) GetProjectWithApiKey(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var project models.Project
	if err := pc.DB.Model(&models.Project{}).
		Where("id = ?", projectID).
		First(&project).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	resultResponse := dtos.GetProjectResponseDetail{
		GetProjectResponse: dtos.GetProjectResponse{
			ProjectID: project.ID.String(),
			Name:      project.Name,
			ApiKey:    project.ApiKey,
			Owner: dtos.OwnerDto{
				ID:    project.Owner.ID.String(),
				Email: project.Owner.Email,
				Name:  project.Owner.Name,
			},
		},
		Devices: project.Devices,
		Events:  project.Events,
	}

	c.JSON(http.StatusOK, resultResponse)
}
