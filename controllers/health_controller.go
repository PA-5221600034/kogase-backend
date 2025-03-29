package controllers

import (
	"net/http"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type HealthController struct {
	DB *gorm.DB
}

func NewHealthController(db *gorm.DB) *HealthController {
	return &HealthController{DB: db}
}

func (h *HealthController) GetHealth(c *gin.Context) {
	response := dtos.HealthResponse{
		Status: "ok",
	}

	c.JSON(http.StatusOK, response)
}

func (h *HealthController) GetHealthWithApiKey(c *gin.Context) {
	// Get project ID from context (set by ApiKeyMiddleware)
	_, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Not authenticated",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	response := dtos.HealthResponse{
		Status: "ok",
	}

	c.JSON(http.StatusOK, response)
}
