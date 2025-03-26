package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/atqamz/kogase-backend/dtos"
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
// @Router /api/v1/analytics/metrics [get]
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

		// Only the owner has access
		if project.OwnerID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Parse query parameters
	var query dtos.MetricsQuery
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
// @Router /api/v1/analytics/events [get]
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

		// Only the owner has access
		if project.OwnerID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Parse query parameters
	var query dtos.EventsQuery
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
// @Router /api/v1/analytics/devices [get]
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

		// Only the owner has access
		if project.OwnerID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
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

// GetRetention returns user retention data
// @Summary Get retention
// @Description Get user retention data
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (ISO 8601)"
// @Param end_date query string false "End date (ISO 8601)"
// @Param cohort_period query string false "Cohort period (daily, weekly, monthly)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/analytics/retention [get]
func (ac *AnalyticsController) GetRetention(c *gin.Context) {
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

		// Only the owner has access
		if project.OwnerID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Parse query parameters
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	cohortPeriod := c.DefaultQuery("cohort_period", "weekly")

	// Set default dates if not provided
	startDate := time.Now().AddDate(0, -3, 0) // 3 months ago
	endDate := time.Now()

	if startDateStr != "" {
		if err := startDate.UnmarshalText([]byte(startDateStr)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
	}

	if endDateStr != "" {
		if err := endDate.UnmarshalText([]byte(endDateStr)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
	}

	// Validate cohort period
	if cohortPeriod != "daily" && cohortPeriod != "weekly" && cohortPeriod != "monthly" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid cohort period"})
		return
	}

	// Implement actual retention calculation
	// We'll get all devices that appeared for the first time in each cohort period
	// Then we'll calculate retention for days 1, 7, 14, 30
	var cohorts []gin.H
	retentionDays := []int{1, 7, 14, 30}

	// Start with the oldest cohort period and move forward
	currentDate := startDate

	// Adjust start date based on cohort period
	switch cohortPeriod {
	case "daily":
		// Keep as is - daily cohorts
	case "weekly":
		// Adjust to start of the week (Monday)
		weekday := int(currentDate.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		currentDate = currentDate.AddDate(0, 0, -(weekday - 1))
	case "monthly":
		// Adjust to start of the month
		currentDate = time.Date(currentDate.Year(), currentDate.Month(), 1, 0, 0, 0, 0, currentDate.Location())
	}

	// Process cohorts until we reach the end date
	for currentDate.Before(endDate) {
		var nextDate time.Time

		// Calculate the next cohort period
		switch cohortPeriod {
		case "daily":
			nextDate = currentDate.AddDate(0, 0, 1)
		case "weekly":
			nextDate = currentDate.AddDate(0, 0, 7)
		case "monthly":
			nextDate = time.Date(currentDate.Year(), currentDate.Month()+1, 1, 0, 0, 0, 0, currentDate.Location())
		}

		// Skip to next period if we've gone beyond end date
		if nextDate.After(endDate) {
			break
		}

		// 1. Find devices that were first seen in this cohort period (new users)
		// We'll use the first_seen field in the devices table
		var newDevices []models.Device
		query := ac.DB.Where("project_id = ? AND first_seen >= ? AND first_seen < ?",
			projectID, currentDate, nextDate)

		if err := query.Find(&newDevices).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate cohort data"})
			return
		}

		// If no new devices in this period, skip to next period
		if len(newDevices) == 0 {
			currentDate = nextDate
			continue
		}

		// Get device IDs for this cohort
		var deviceIDs []uuid.UUID
		for _, device := range newDevices {
			deviceIDs = append(deviceIDs, device.ID)
		}

		// Prepare retention data structure
		newUserCount := len(newDevices)
		retentionData := []gin.H{
			{"day": 0, "users": newUserCount, "percentage": 100}, // Day 0 is always 100%
		}

		// 2. For each retention period (day 1, 7, 14, etc.), find how many devices returned
		for _, day := range retentionDays {
			periodStart := nextDate // End of the cohort period
			periodEnd := periodStart.AddDate(0, 0, day)

			// Count active devices from the cohort during this retention period
			var activeCount int64
			activeQuery := ac.DB.Model(&models.Event{}).
				Where("project_id = ? AND device_id IN ? AND timestamp >= ? AND timestamp < ?",
					projectID, deviceIDs, periodStart, periodEnd).
				Distinct("device_id").Count(&activeCount)

			if activeQuery.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate retention data"})
				return
			}

			// Calculate percentage
			percentage := 0.0
			if newUserCount > 0 {
				percentage = float64(activeCount) / float64(newUserCount) * 100
			}

			retentionData = append(retentionData, gin.H{
				"day":        day,
				"users":      activeCount,
				"percentage": percentage,
			})
		}

		// Add this cohort to the result
		cohorts = append(cohorts, gin.H{
			"cohort_date": currentDate.Format("2006-01-02"),
			"new_users":   newUserCount,
			"retention":   retentionData,
		})

		// Move to next period
		currentDate = nextDate
	}

	// If no cohorts were found, provide an empty result
	if len(cohorts) == 0 {
		cohorts = []gin.H{}
	}

	c.JSON(http.StatusOK, gin.H{
		"cohorts": cohorts,
	})
}

// GetSessions returns session data
// @Summary Get sessions
// @Description Get session data
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (ISO 8601)"
// @Param end_date query string false "End date (ISO 8601)"
// @Param platform query string false "Platform"
// @Param limit query int false "Limit (default 100)"
// @Param offset query int false "Offset (default 0)"
// @Security BearerAuth
// @Success 200 {array} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/analytics/sessions [get]
func (ac *AnalyticsController) GetSessions(c *gin.Context) {
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

		// Only the owner has access
		if project.OwnerID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Parse query parameters
	platform := c.Query("platform")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	// Set default dates if not provided
	startDate := time.Now().AddDate(0, -1, 0) // 1 month ago
	endDate := time.Now()

	if startDateStr != "" {
		if err := startDate.UnmarshalText([]byte(startDateStr)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
	}

	if endDateStr != "" {
		if err := endDate.UnmarshalText([]byte(endDateStr)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
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

	// Query for session start events
	dbQuery := ac.DB.Model(&models.Event{}).
		Select("events.*, devices.device_id as client_device_id, devices.platform, devices.os_version, devices.app_version").
		Joins("JOIN devices ON events.device_id = devices.id").
		Where("events.project_id = ? AND events.event_type = ?", projectID, models.SessionStart).
		Where("events.timestamp BETWEEN ? AND ?", startDate, endDate)

	if platform != "" {
		dbQuery = dbQuery.Where("devices.platform = ?", platform)
	}

	// Apply limit and offset
	dbQuery = dbQuery.Limit(limit).Offset(offset).Order("events.timestamp DESC")

	// Execute query
	rows, err := dbQuery.Rows()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve sessions"})
		return
	}
	defer rows.Close()

	// Process results into sessions
	var sessions []gin.H
	for rows.Next() {
		var event models.Event
		var deviceID string
		var platform string
		var osVersion string
		var appVersion string

		if err := ac.DB.ScanRows(rows, &event); err != nil {
			continue
		}

		if err := rows.Scan(nil, nil, nil, nil, nil, nil, nil, nil, &deviceID, &platform, &osVersion, &appVersion); err != nil {
			continue
		}

		// For each session start, try to find corresponding session end
		var sessionEnd models.Event
		endResult := ac.DB.Where("project_id = ? AND device_id = ? AND event_type = ? AND timestamp > ?",
			projectID, event.DeviceID, models.SessionEnd, event.Timestamp).
			Order("timestamp ASC").Limit(1).Find(&sessionEnd)

		var duration float64
		var sessionEndTime *time.Time

		if endResult.RowsAffected > 0 {
			duration = sessionEnd.Timestamp.Sub(event.Timestamp).Seconds()
			sessionEndTime = &sessionEnd.Timestamp
		}

		sessions = append(sessions, gin.H{
			"session_id":  event.ID.String(),
			"device_id":   deviceID,
			"platform":    platform,
			"os_version":  osVersion,
			"app_version": appVersion,
			"start_time":  event.Timestamp,
			"end_time":    sessionEndTime,
			"duration":    duration,
		})
	}

	c.JSON(http.StatusOK, sessions)
}

// GetActiveUsers returns daily and monthly active user data
// @Summary Get active users
// @Description Get daily and monthly active user counts
// @Tags analytics
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (ISO 8601)"
// @Param end_date query string false "End date (ISO 8601)"
// @Param platform query string false "Platform"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/analytics/active-users [get]
func (ac *AnalyticsController) GetActiveUsers(c *gin.Context) {
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

		// Only the owner has access
		if project.OwnerID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			return
		}
	}

	// Parse query parameters
	platform := c.Query("platform")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	// Set default dates if not provided
	startDate := time.Now().AddDate(0, -1, 0) // 1 month ago
	endDate := time.Now()

	if startDateStr != "" {
		if err := startDate.UnmarshalText([]byte(startDateStr)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
	}

	if endDateStr != "" {
		if err := endDate.UnmarshalText([]byte(endDateStr)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
	}

	// Calculate daily active users (DAU)
	var dailyActiveUsers []gin.H
	currentDate := startDate
	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		nextDate := currentDate.AddDate(0, 0, 1)

		// Build query for distinct device count
		deviceQuery := ac.DB.Model(&models.Event{}).
			Joins("JOIN devices ON events.device_id = devices.id").
			Where("events.project_id = ?", projectID).
			Where("events.timestamp >= ? AND events.timestamp < ?", currentDate, nextDate)

		if platform != "" {
			deviceQuery = deviceQuery.Where("devices.platform = ?", platform)
		}

		// Count distinct devices
		var distinctCount int64
		if err := deviceQuery.Distinct("events.device_id").Count(&distinctCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate DAU"})
			return
		}

		dailyActiveUsers = append(dailyActiveUsers, gin.H{
			"date":  currentDate.Format("2006-01-02"),
			"count": distinctCount,
		})

		currentDate = nextDate
	}

	// Calculate monthly active users (MAU)
	var monthlyActiveUsers []gin.H

	// Reset to start of month
	currentDate = time.Date(startDate.Year(), startDate.Month(), 1, 0, 0, 0, 0, startDate.Location())

	for currentDate.Before(endDate) || currentDate.Equal(endDate) {
		// Calculate next month
		nextMonth := currentDate.AddDate(0, 1, 0)

		// Build query for distinct device count
		deviceQuery := ac.DB.Model(&models.Event{}).
			Joins("JOIN devices ON events.device_id = devices.id").
			Where("events.project_id = ?", projectID).
			Where("events.timestamp >= ? AND events.timestamp < ?", currentDate, nextMonth)

		if platform != "" {
			deviceQuery = deviceQuery.Where("devices.platform = ?", platform)
		}

		// Count distinct devices
		var distinctCount int64
		if err := deviceQuery.Distinct("events.device_id").Count(&distinctCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate MAU"})
			return
		}

		monthlyActiveUsers = append(monthlyActiveUsers, gin.H{
			"month": currentDate.Format("2006-01"),
			"count": distinctCount,
		})

		currentDate = nextMonth
	}

	// Calculate current DAU and MAU
	today := time.Now()
	yesterday := today.AddDate(0, 0, -1)
	startOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())

	// DAU query
	dauQuery := ac.DB.Model(&models.Event{}).
		Joins("JOIN devices ON events.device_id = devices.id").
		Where("events.project_id = ?", projectID).
		Where("events.timestamp >= ? AND events.timestamp < ?", yesterday, today)

	// MAU query
	mauQuery := ac.DB.Model(&models.Event{}).
		Joins("JOIN devices ON events.device_id = devices.id").
		Where("events.project_id = ?", projectID).
		Where("events.timestamp >= ? AND events.timestamp < ?", startOfMonth, today.AddDate(0, 0, 1))

	if platform != "" {
		dauQuery = dauQuery.Where("devices.platform = ?", platform)
		mauQuery = mauQuery.Where("devices.platform = ?", platform)
	}

	// Current counts
	var currentDAU, currentMAU int64

	if err := dauQuery.Distinct("events.device_id").Count(&currentDAU).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate current DAU"})
		return
	}

	if err := mauQuery.Distinct("events.device_id").Count(&currentMAU).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate current MAU"})
		return
	}

	// Create response
	c.JSON(http.StatusOK, gin.H{
		"daily_active_users":   dailyActiveUsers,
		"monthly_active_users": monthlyActiveUsers,
		"current": gin.H{
			"dau": currentDAU,
			"mau": currentMAU,
			"ratio": func() float64 {
				if currentMAU == 0 {
					return 0
				}
				return float64(currentDAU) / float64(currentMAU)
			}(),
		},
	})
}
