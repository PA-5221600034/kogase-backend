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

// GetHealth godoc
// @Summary Health check endpoint
// @Description Check if the API is running
// @Tags health
// @Produce json
// @Success 200 {object} dtos.HealthResponse
// @Router /health [get]
func (h *HealthController) GetHealth(c *gin.Context) {
	resultResponse := dtos.HealthResponse{
		Status: "ok",
	}

	c.JSON(http.StatusOK, resultResponse)
}

// GetHealthWithApiKey godoc
// @Summary Health check endpoint with API key authentication
// @Description Check if the API is running and verify API key
// @Tags health
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dtos.HealthResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Router /health/apikey [get]
func (h *HealthController) GetHealthWithApiKey(c *gin.Context) {
	_, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Not authenticated",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	resultResponse := dtos.HealthResponse{
		Status: "ok",
	}

	c.JSON(http.StatusOK, resultResponse)
}
