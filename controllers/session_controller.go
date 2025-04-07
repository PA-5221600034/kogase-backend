package controllers

import (
	"net/http"
	"time"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SessionController struct {
	DB *gorm.DB
}

func NewSessionController(db *gorm.DB) *SessionController {
	return &SessionController{DB: db}
}

func (sc *SessionController) BeginSession(c *gin.Context) {
	// Get project ID from context (set by ApiKeyMiddleware)
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Bind request body
	var request dtos.BeginSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request body",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate device exists
	device := models.Device{
		Identifier: request.Identifier,
	}
	if err := sc.DB.First(&device, "identifier = ?", request.Identifier).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Device not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Create session
	session := models.Session{
		ProjectID: projectID.(uuid.UUID),
		DeviceID:  device.ID,
	}
	if err := sc.DB.Create(&session).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to create session",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Return session ID
	response := dtos.BeginSessionResponse{
		SessionID: session.ID.String(),
	}

	// Return response
	c.JSON(http.StatusCreated, response)
}

func (sc *SessionController) EndSession(c *gin.Context) {
	// Get project ID from context (set by ApiKeyMiddleware)
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Bind request body
	var request dtos.EndSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request body",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate session exists
	session := models.Session{
		ID:        uuid.MustParse(request.SessionID),
		ProjectID: projectID.(uuid.UUID),
	}
	if err := sc.DB.First(&session, "id = ? AND project_id = ?", request.SessionID, projectID.(uuid.UUID)).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Session not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Update session finish time
	session.EndAt = time.Now()
	session.Duration = time.Since(session.BeginAt)
	if err := sc.DB.Save(&session).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to end session",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Return session ID
	response := dtos.EndSessionResponse{
		Message: "Session ended",
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

func (sc *SessionController) GetDeviceSessions(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Bind request query
	var request dtos.GetDeviceSessionsRequestQuery
	if err := c.ShouldBindQuery(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request query",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Validate device exists
	device := models.Device{
		ID: request.DeviceID,
	}
	if err := sc.DB.First(&device, "id = ?", request.DeviceID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Device not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Get sessions
	var sessions []models.Session
	if err := sc.DB.Where("project_id = ? AND device_id = ?", request.ProjectID, request.DeviceID).Find(&sessions).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to get sessions",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Convert sessions to dtos
	var sessionsDTO []dtos.GetSessionResponse
	for _, session := range sessions {
		sessionsDTO = append(sessionsDTO, dtos.GetSessionResponse{
			SessionID: session.ID,
			ProjectID: session.ProjectID,
			DeviceID:  session.DeviceID,
			BeginAt:   session.BeginAt,
			EndAt:     session.EndAt,
		})
	}

	// Return sessions
	response := dtos.GetSessionsResponse{
		Sessions: sessionsDTO,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

func (sc *SessionController) GetProjectSessions(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Bind request query
	var request dtos.GetProjectSessionsRequestQuery
	if err := c.ShouldBindQuery(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request query",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get sessions
	var sessions []models.Session
	if err := sc.DB.Where("project_id = ?", request.ProjectID).Find(&sessions).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to get sessions",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Convert sessions to dtos
	var sessionsDTO []dtos.GetSessionResponse
	for _, session := range sessions {
		sessionsDTO = append(sessionsDTO, dtos.GetSessionResponse{
			SessionID: session.ID,
			ProjectID: session.ProjectID,
			DeviceID:  session.DeviceID,
			BeginAt:   session.BeginAt,
			EndAt:     session.EndAt,
		})
	}

	// Return sessions
	response := dtos.GetSessionsResponse{
		Sessions: sessionsDTO,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

func (sc *SessionController) GetAllSessions(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Bind request query
	var request dtos.GetAllSessionsRequestQuery
	if err := c.ShouldBindQuery(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request query",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get sessions
	var sessions []models.Session
	if err := sc.DB.Find(&sessions).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to get sessions",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Convert sessions to dtos
	var sessionsDTO []dtos.GetSessionResponse
	for _, session := range sessions {
		sessionsDTO = append(sessionsDTO, dtos.GetSessionResponse{
			SessionID: session.ID,
		})
	}

	// Return sessions
	response := dtos.GetSessionsResponse{
		Sessions: sessionsDTO,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}
