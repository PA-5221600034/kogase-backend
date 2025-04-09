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
		ProjectID: project.ID,
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID,
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
	}

	c.JSON(http.StatusCreated, resultResponse)
}

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

	c.JSON(http.StatusOK, resultResponse)
}

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
			ProjectID: project.ID,
			Name:      project.Name,
			ApiKey:    project.ApiKey,
			Owner: dtos.OwnerDto{
				ID:    project.Owner.ID,
				Email: project.Owner.Email,
				Name:  project.Owner.Name,
			},
		},
		Devices: project.Devices,
		Events:  project.Events,
	}

	c.JSON(http.StatusOK, resultResponse)
}

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
		ProjectID: project.ID,
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID,
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
	}

	c.JSON(http.StatusOK, resultResponse)
}

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
		ProjectID: project.ID,
		Name:      project.Name,
		ApiKey:    project.ApiKey,
		Owner: dtos.OwnerDto{
			ID:    project.Owner.ID,
			Email: project.Owner.Email,
			Name:  project.Owner.Name,
		},
	}

	c.JSON(http.StatusOK, resultResponse)
}

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
			ProjectID: project.ID,
			Name:      project.Name,
			ApiKey:    project.ApiKey,
			Owner: dtos.OwnerDto{
				ID:    project.Owner.ID,
				Email: project.Owner.Email,
				Name:  project.Owner.Name,
			},
		},
		Devices: project.Devices,
		Events:  project.Events,
	}

	c.JSON(http.StatusOK, resultResponse)
}
