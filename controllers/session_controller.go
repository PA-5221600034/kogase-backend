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

func (sc *SessionController) GetSessions(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get user
	var user models.User
	if err := sc.DB.First(&user, "id = ?", userID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Bind request query
	var request dtos.GetSessionsRequestQuery
	if err := c.ShouldBindQuery(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request query",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	query := sc.DB

	if request.ProjectID != uuid.Nil {
		query = query.Where("project_id = ?", request.ProjectID)
	}

	if !request.FromDate.IsZero() {
		fromDate := time.Date(
			request.FromDate.Year(),
			request.FromDate.Month(),
			request.FromDate.Day()+1,
			0, 0, 0, 0,
			request.FromDate.Location(),
		)
		query = query.Where("begin_at >= ?", fromDate)
	}

	if !request.ToDate.IsZero() {
		toDate := time.Date(
			request.ToDate.Year(),
			request.ToDate.Month(),
			request.ToDate.Day()+1,
			23, 59, 59, 999999999,
			request.ToDate.Location(),
		)
		query = query.Where("end_at <= ?", toDate)
	}

	// Pagination
	query = query.Limit(request.Limit).Offset(request.Offset)

	// Order
	query = query.Order("begin_at DESC")

	var sessions []models.Session
	if err := query.Find(&sessions).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to get sessions",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	if len(sessions) == 0 {
		response := dtos.GetSessionsResponse{
			Sessions: []dtos.GetSessionResponse{},
			Total:    0,
			Limit:    request.Limit,
			Offset:   request.Offset,
		}
		c.JSON(http.StatusOK, response)
		return
	}

	// Convert sessions to dtos
	var sessionsDTO []dtos.GetSessionResponse
	for _, session := range sessions {
		sessionsDTO = append(sessionsDTO, dtos.GetSessionResponse{
			SessionID: session.ID,
			BeginAt:   session.BeginAt,
			EndAt:     session.EndAt,
			Duration:  session.Duration,
		})
	}

	// Return sessions
	response := dtos.GetSessionsResponse{
		Sessions: sessionsDTO,
		Total:    int64(len(sessions)),
		Limit:    request.Limit,
		Offset:   request.Offset,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

func (sc *SessionController) GetSession(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get session ID from path
	sessionID := c.Param("id")

	// Validate session exists
	session := models.Session{
		ID: uuid.MustParse(sessionID),
	}
	if err := sc.DB.First(&session, "id = ?", sessionID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Session not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Return session
	response := dtos.GetSessionResponse{
		SessionID: session.ID,
		BeginAt:   session.BeginAt,
		EndAt:     session.EndAt,
		Duration:  session.Duration,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}
