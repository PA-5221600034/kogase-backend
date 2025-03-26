package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/atqamz/kogase-backend/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AnalyticsController handles analytics-related endpoints
type AnalyticsController struct {
	DB *gorm.DB
}

// NewAnalyticsController creates a new AnalyticsController instance
func NewAnalyticsController(db *gorm.DB) *AnalyticsController {
	return &AnalyticsController{DB: db}
}

// MetricsQuery represents a query for analytics metrics
type MetricsQuery struct {
	MetricType string     `form:"metric_type"`
	StartDate  *time.Time `form:"start_date"`
	EndDate    *time.Time `form:"end_date"`
	Period     string     `form:"period" binding:"omitempty,oneof=hourly daily weekly monthly yearly total"`
	Dimensions []string   `form:"dimensions"`
}

// GetMetrics returns analytics metrics
// @Summary Get metrics
// @Description Get analytics metrics
// @Tags analytics
// @Accept json
// @Produce json
// @Param metric_type query string false "Metric type"
// @Param start_date query string false "Start date (ISO 8601)"
// @Param end_date query string false "End date (ISO 8601)"
// @Param period query string false "Period (hourly, daily, weekly, monthly, yearly, total)"
// @Param dimensions query string false "Dimensions to group by"
// @Security BearerAuth
// @Success 200 {array} models.Metric
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /analytics/metrics [get]
func (ac *AnalyticsController) GetMetrics(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		// Check if project ID is provided in query
		projectIDStr := c.Query("project_id")
		if projectIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
			return
		}

		// Parse project ID
		var err error
		projectID, err = uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}

		// Check if user has access to project
		var project models.Project
		if err := ac.DB.First(&project, "id = ?", projectID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}

		role, _ := c.Get("user_role")
		if role != "admin" && project.OwnerID != userID.(uuid.UUID) {
			var projectUser models.ProjectUser
			if err := ac.DB.Where("project_id = ? AND user_id = ?", projectID, userID).First(&projectUser).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}
	}

	// Parse query parameters
	var query MetricsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	// Build query
	dbQuery := ac.DB.Where("project_id = ?", projectID)

	if query.MetricType != "" {
		dbQuery = dbQuery.Where("metric_type = ?", query.MetricType)
	}

	if query.StartDate != nil {
		dbQuery = dbQuery.Where("period_start >= ?", query.StartDate)
	}

	if query.EndDate != nil {
		dbQuery = dbQuery.Where("period_start <= ?", query.EndDate)
	}

	if query.Period != "" {
		dbQuery = dbQuery.Where("period = ?", query.Period)
	}

	// Get metrics
	var metrics []models.Metric
	if err := dbQuery.Find(&metrics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve metrics"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// EventsQuery represents a query for events
type EventsQuery struct {
	EventType string     `form:"event_type"`
	EventName string     `form:"event_name"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
	DeviceID  string     `form:"device_id"`
	Platform  string     `form:"platform"`
	Limit     int        `form:"limit,default=100"`
	Offset    int        `form:"offset,default=0"`
}

// GetEvents returns events
// @Summary Get events
// @Description Get events
// @Tags analytics
// @Accept json
// @Produce json
// @Param event_type query string false "Event type"
// @Param event_name query string false "Event name"
// @Param start_date query string false "Start date (ISO 8601)"
// @Param end_date query string false "End date (ISO 8601)"
// @Param device_id query string false "Device ID"
// @Param platform query string false "Platform"
// @Param limit query int false "Limit (default 100)"
// @Param offset query int false "Offset (default 0)"
// @Security BearerAuth
// @Success 200 {array} models.Event
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /analytics/events [get]
func (ac *AnalyticsController) GetEvents(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		// Check if project ID is provided in query
		projectIDStr := c.Query("project_id")
		if projectIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
			return
		}

		// Parse project ID
		var err error
		projectID, err = uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}

		// Check if user has access to project
		var project models.Project
		if err := ac.DB.First(&project, "id = ?", projectID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}

		role, _ := c.Get("user_role")
		if role != "admin" && project.OwnerID != userID.(uuid.UUID) {
			var projectUser models.ProjectUser
			if err := ac.DB.Where("project_id = ? AND user_id = ?", projectID, userID).First(&projectUser).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}
	}

	// Parse query parameters
	var query EventsQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid query parameters"})
		return
	}

	// Build query
	dbQuery := ac.DB.Where("project_id = ?", projectID)

	if query.EventType != "" {
		dbQuery = dbQuery.Where("event_type = ?", query.EventType)
	}

	if query.EventName != "" {
		dbQuery = dbQuery.Where("event_name = ?", query.EventName)
	}

	if query.StartDate != nil {
		dbQuery = dbQuery.Where("timestamp >= ?", query.StartDate)
	}

	if query.EndDate != nil {
		dbQuery = dbQuery.Where("timestamp <= ?", query.EndDate)
	}

	if query.DeviceID != "" {
		var device models.Device
		if err := ac.DB.Where("project_id = ? AND device_id = ?", projectID, query.DeviceID).First(&device).Error; err == nil {
			dbQuery = dbQuery.Where("device_id = ?", device.ID)
		} else {
			// If device not found, return empty result
			c.JSON(http.StatusOK, []models.Event{})
			return
		}
	}

	if query.Platform != "" {
		dbQuery = dbQuery.Joins("JOIN devices ON events.device_id = devices.id").
			Where("devices.platform = ?", query.Platform)
	}

	// Apply limit and offset
	if query.Limit <= 0 {
		query.Limit = 100
	} else if query.Limit > 1000 {
		query.Limit = 1000
	}

	dbQuery = dbQuery.Limit(query.Limit).Offset(query.Offset)

	// Get events
	var events []models.Event
	if err := dbQuery.Order("timestamp DESC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// GetDevices returns devices
// @Summary Get devices
// @Description Get devices
// @Tags analytics
// @Accept json
// @Produce json
// @Param platform query string false "Platform"
// @Param start_date query string false "First seen date (ISO 8601)"
// @Param end_date query string false "Last seen date (ISO 8601)"
// @Param limit query int false "Limit (default 100)"
// @Param offset query int false "Offset (default 0)"
// @Security BearerAuth
// @Success 200 {array} models.Device
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /analytics/devices [get]
func (ac *AnalyticsController) GetDevices(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			return
		}

		// Check if project ID is provided in query
		projectIDStr := c.Query("project_id")
		if projectIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
			return
		}

		// Parse project ID
		var err error
		projectID, err = uuid.Parse(projectIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
			return
		}

		// Check if user has access to project
		var project models.Project
		if err := ac.DB.First(&project, "id = ?", projectID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}

		role, _ := c.Get("user_role")
		if role != "admin" && project.OwnerID != userID.(uuid.UUID) {
			var projectUser models.ProjectUser
			if err := ac.DB.Where("project_id = ? AND user_id = ?", projectID, userID).First(&projectUser).Error; err != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
				return
			}
		}
	}

	// Parse query parameters
	platform := c.Query("platform")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	// Build query
	dbQuery := ac.DB.Where("project_id = ?", projectID)

	if platform != "" {
		dbQuery = dbQuery.Where("platform = ?", platform)
	}

	if startDateStr != "" {
		var startDate time.Time
		if err := startDate.UnmarshalText([]byte(startDateStr)); err == nil {
			dbQuery = dbQuery.Where("first_seen >= ?", startDate)
		}
	}

	if endDateStr != "" {
		var endDate time.Time
		if err := endDate.UnmarshalText([]byte(endDateStr)); err == nil {
			dbQuery = dbQuery.Where("last_seen <= ?", endDate)
		}
	}

	// Parse limit and offset
	var limit int
	var offset int
	if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil || limit <= 0 {
		limit = 100
	} else if limit > 1000 {
		limit = 1000
	}

	if _, err := fmt.Sscanf(offsetStr, "%d", &offset); err != nil || offset < 0 {
		offset = 0
	}

	// Apply limit and offset
	dbQuery = dbQuery.Limit(limit).Offset(offset)

	// Get devices
	var devices []models.Device
	if err := dbQuery.Order("last_seen DESC").Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve devices"})
		return
	}

	c.JSON(http.StatusOK, devices)
}
