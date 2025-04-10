package controllers

import (
	"net/http"
	"time"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AnalyticsController struct {
	DB *gorm.DB
}

func NewAnalyticsController(db *gorm.DB) *AnalyticsController {
	return &AnalyticsController{DB: db}
}

func (ac *AnalyticsController) GetAnalytics(c *gin.Context) {
	_, exist := c.Get("user_id")
	if !exist {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
	}

	var request dtos.GetAnalyticsRequestQuery
	if err := c.ShouldBindQuery(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request query",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	sessionQuery := ac.DB.Model(&models.Session{})
	if request.ProjectID != "" {
		sessionQuery = sessionQuery.Where("project_id = ?", request.ProjectID)
	}
	if !request.FromDate.IsZero() {
		sessionQuery = sessionQuery.Where("begin_at >= ?", request.FromDate)
	}
	if !request.ToDate.IsZero() {
		sessionQuery = sessionQuery.Where("begin_at <= ?", request.ToDate)
	}

	response := dtos.GetAnalyticsResponse{
		DAU:           0,
		MAU:           0,
		TotalDuration: 0,
		TotalInstalls: 0,
	}

	var sessions []models.Session
	if err := sessionQuery.Find(&sessions).Error; err != nil {
		response.DAU = 0
		response.MAU = 0
		response.TotalDuration = 0
	} else {
		for _, session := range sessions {
			if session.BeginAt.After(time.Now().AddDate(0, 0, -1)) {
				response.DAU++
			}
		}
		for _, session := range sessions {
			if session.BeginAt.After(time.Now().AddDate(0, 0, -30)) {
				response.MAU++
			}
		}
		for _, session := range sessions {
			response.TotalDuration += session.Duration
		}
	}

	eventQuery := ac.DB.Model(&models.Event{})
	if request.ProjectID != "" {
		eventQuery = eventQuery.Where("project_id = ?", request.ProjectID)
	}
	if !request.FromDate.IsZero() {
		eventQuery = eventQuery.Where("received_at >= ?", request.FromDate)
	}
	if !request.ToDate.IsZero() {
		eventQuery = eventQuery.Where("received_at <= ?", request.ToDate)
	}

	var totalInstalls int64
	if err := eventQuery.Model(&models.Event{}).
		Where("event_type = ? AND event_name = ?", "predefined", "install").
		Count(&totalInstalls).Error; err != nil {
		response.TotalInstalls = 0
	} else {
		response.TotalInstalls = int(totalInstalls)
	}

	c.JSON(http.StatusOK, response)
}
