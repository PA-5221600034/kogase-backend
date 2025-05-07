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

// BeginSession godoc
// @Summary Begin a new session
// @Description Start a new game session for a device
// @Tags sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param session body dtos.BeginSessionRequest true "Session details"
// @Success 201 {object} dtos.BeginSessionResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /sessions/begin [post]
func (sc *SessionController) BeginSession(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var request dtos.BeginSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request body",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	device := models.Device{
		Identifier: request.Identifier,
	}
	if err := sc.DB.Model(&models.Device{}).
		Where("identifier = ?", request.Identifier).
		First(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Device not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

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

	resultResponse := dtos.BeginSessionResponse{
		SessionID: session.ID.String(),
	}

	c.JSON(http.StatusCreated, resultResponse)
}

// EndSession godoc
// @Summary End a session
// @Description End an existing game session
// @Tags sessions
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param session body dtos.EndSessionRequest true "Session to end"
// @Success 200 {object} dtos.EndSessionResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /sessions/end [post]
func (sc *SessionController) EndSession(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var request dtos.EndSessionRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request body",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	session := models.Session{
		ID:        uuid.MustParse(request.SessionID),
		ProjectID: projectID.(uuid.UUID),
	}
	if err := sc.DB.Model(&models.Session{}).
		Where("id = ? AND project_id = ?", request.SessionID, projectID.(uuid.UUID)).
		First(&session).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Session not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	session.EndAt = time.Now()
	session.Duration = time.Since(session.BeginAt)
	if err := sc.DB.Save(&session).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to end session",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.EndSessionResponse{
		Message: "Session ended",
	}

	c.JSON(http.StatusOK, resultResponse)
}

// GetSessions godoc
// @Summary Get sessions
// @Description Retrieve all sessions with filtering and pagination
// @Tags sessions
// @Produce json
// @Security BearerAuth
// @Param project_id query string false "Filter by project ID"
// @Param from_date query string false "Filter by start date (RFC3339)"
// @Param to_date query string false "Filter by end date (RFC3339)"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} dtos.GetSessionsResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /sessions [get]
func (sc *SessionController) GetSessions(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var request dtos.GetSessionsRequestQuery
	if err := c.ShouldBindQuery(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request query",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	query := sc.DB.Model(&models.Session{})
	if request.ProjectID != "" {
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

	query = query.Limit(request.Limit).Offset(request.Offset)

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

	var sessionsDTO []dtos.GetSessionResponse
	for _, session := range sessions {
		sessionsDTO = append(sessionsDTO, dtos.GetSessionResponse{
			SessionID: session.ID.String(),
			BeginAt:   session.BeginAt,
			EndAt:     session.EndAt,
			Duration:  session.Duration.Nanoseconds(),
		})
	}

	resultResponse := dtos.GetSessionsResponse{
		Sessions: sessionsDTO,
		Total:    int64(len(sessions)),
		Limit:    request.Limit,
		Offset:   request.Offset,
	}

	c.JSON(http.StatusOK, resultResponse)
}

// GetSession godoc
// @Summary Get session by ID
// @Description Retrieve a specific session by its ID
// @Tags sessions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 200 {object} dtos.GetSessionResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /sessions/{id} [get]
func (sc *SessionController) GetSession(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	sessionID := c.Param("id")
	session := models.Session{
		ID: uuid.MustParse(sessionID),
	}
	if err := sc.DB.Model(&models.Session{}).
		Where("id = ?", sessionID).
		First(&session).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Session not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	resultResponse := dtos.GetSessionResponse{
		SessionID: session.ID.String(),
		BeginAt:   session.BeginAt,
		EndAt:     session.EndAt,
		Duration:  session.Duration.Nanoseconds(),
	}

	c.JSON(http.StatusOK, resultResponse)
}
