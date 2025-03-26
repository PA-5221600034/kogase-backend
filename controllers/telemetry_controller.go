package controllers

import (
	"net/http"
	"time"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TelemetryController handles telemetry-related endpoints
type TelemetryController struct {
	DB *gorm.DB
}

// NewTelemetryController creates a new TelemetryController instance
func NewTelemetryController(db *gorm.DB) *TelemetryController {
	return &TelemetryController{DB: db}
}

// findOrCreateDevice finds an existing device or creates a new one
func (tc *TelemetryController) findOrCreateDevice(c *gin.Context, projectID uuid.UUID, deviceID string, platform string, osVersion string, appVersion string) (models.Device, error) {
	var device models.Device

	// Try to find the device
	result := tc.DB.Where("project_id = ? AND device_id = ?", projectID, deviceID).First(&device)

	// Current IP address
	ipAddress := c.ClientIP()

	// Device found - update it
	if result.Error == nil {
		device.LastSeen = time.Now()

		// Only update these if provided
		if appVersion != "" {
			device.AppVersion = appVersion
		}

		if osVersion != "" {
			device.OSVersion = osVersion
		}

		// Update IP and country if they've changed
		if device.IPAddress != ipAddress {
			device.IPAddress = ipAddress

			// Get updated country from IP address
			country, err := utils.GetCountryFromIP(ipAddress)
			if err == nil && country != "Unknown" {
				device.Country = country
			}
		}

		if err := tc.DB.Save(&device).Error; err != nil {
			return device, err
		}

		return device, nil
	}

	// Device doesn't exist, create a new one
	// Get country from IP address
	country, err := utils.GetCountryFromIP(ipAddress)
	if err != nil {
		// If geolocation fails, continue with unknown country
		country = "Unknown"
	}

	// Create new device
	newDevice := models.Device{
		ProjectID:  projectID,
		DeviceID:   deviceID,
		Platform:   platform,
		OSVersion:  osVersion,
		AppVersion: appVersion,
		FirstSeen:  time.Now(),
		LastSeen:   time.Now(),
		IPAddress:  ipAddress,
		Country:    country,
	}

	if err := tc.DB.Create(&newDevice).Error; err != nil {
		return newDevice, err
	}

	return newDevice, nil
}

// RecordEvent records a single telemetry event
// @Summary Record event
// @Description Record a telemetry event
// @Tags telemetry
// @Accept json
// @Produce json
// @Param event body dtos.EventRequest true "Event details"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/telemetry/events [post]
func (tc *TelemetryController) RecordEvent(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}

	var eventReq dtos.EventRequest
	if err := c.ShouldBindJSON(&eventReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find or create device
	device, err := tc.findOrCreateDevice(
		c,
		projectID.(uuid.UUID),
		eventReq.DeviceID,
		eventReq.Platform,
		eventReq.OSVersion,
		eventReq.AppVersion,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create or update device"})
		return
	}

	// Create event
	timestamp := time.Now()
	if eventReq.Timestamp != nil {
		timestamp = *eventReq.Timestamp
	}

	event := models.Event{
		ProjectID:  projectID.(uuid.UUID),
		DeviceID:   device.ID,
		EventType:  models.EventType(eventReq.EventType),
		EventName:  eventReq.EventName,
		Parameters: eventReq.Parameters,
		Timestamp:  timestamp,
		ReceivedAt: time.Now(),
	}

	if err := tc.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Event recorded successfully"})
}

// RecordEvents records multiple telemetry events in a batch
// @Summary Record events batch
// @Description Record multiple telemetry events in a batch
// @Tags telemetry
// @Accept json
// @Produce json
// @Param events body dtos.EventsRequest true "Events batch"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/telemetry/events/batch [post]
func (tc *TelemetryController) RecordEvents(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}

	var eventsReq dtos.EventsRequest
	if err := c.ShouldBindJSON(&eventsReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Process events in a transaction
	err := tc.DB.Transaction(func(tx *gorm.DB) error {
		// Keep track of devices to avoid multiple lookups
		devices := make(map[string]models.Device)

		// Current IP address - get once for efficiency
		ipAddress := c.ClientIP()

		// Get country from IP - do it once for all devices in the batch
		country, err := utils.GetCountryFromIP(ipAddress)
		if err != nil {
			country = "Unknown"
		}

		for _, eventReq := range eventsReq.Events {
			// Find or create device
			device, exists := devices[eventReq.DeviceID]

			if !exists {
				var dbDevice models.Device
				deviceResult := tx.Where("project_id = ? AND device_id = ?", projectID, eventReq.DeviceID).First(&dbDevice)

				if deviceResult.Error != nil {
					// Device doesn't exist, create a new one
					dbDevice = models.Device{
						ProjectID:  projectID.(uuid.UUID),
						DeviceID:   eventReq.DeviceID,
						Platform:   eventReq.Platform,
						OSVersion:  eventReq.OSVersion,
						AppVersion: eventReq.AppVersion,
						FirstSeen:  time.Now(),
						LastSeen:   time.Now(),
						IPAddress:  ipAddress,
						Country:    country,
					}

					if err := tx.Create(&dbDevice).Error; err != nil {
						return err
					}
				} else {
					// Update device
					dbDevice.LastSeen = time.Now()
					dbDevice.AppVersion = eventReq.AppVersion
					dbDevice.OSVersion = eventReq.OSVersion

					// Update IP and country if they've changed
					if dbDevice.IPAddress != ipAddress {
						dbDevice.IPAddress = ipAddress
						dbDevice.Country = country
					}

					if err := tx.Save(&dbDevice).Error; err != nil {
						return err
					}
				}

				devices[eventReq.DeviceID] = dbDevice
				device = dbDevice
			}

			// Create event
			timestamp := time.Now()
			if eventReq.Timestamp != nil {
				timestamp = *eventReq.Timestamp
			}

			event := models.Event{
				ProjectID:  projectID.(uuid.UUID),
				DeviceID:   device.ID,
				EventType:  models.EventType(eventReq.EventType),
				EventName:  eventReq.EventName,
				Parameters: eventReq.Parameters,
				Timestamp:  timestamp,
				ReceivedAt: time.Now(),
			}

			if err := tx.Create(&event).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record events"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Events recorded successfully", "count": len(eventsReq.Events)})
}

// StartSession starts a new session for a device
// @Summary Start session
// @Description Start a new session for a device
// @Tags telemetry
// @Accept json
// @Produce json
// @Param session body dtos.EventRequest true "Session details"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/telemetry/session/start [post]
func (tc *TelemetryController) StartSession(c *gin.Context) {
	// Session start is essentially an event with type "session_start"
	c.Request.Body = http.NoBody // Clear existing body to avoid binding issues
	c.Request.Header.Set("Content-Type", "application/json")

	var eventReq dtos.EventRequest
	if err := c.ShouldBindJSON(&eventReq); err != nil {
		eventReq = dtos.EventRequest{
			DeviceID:   c.Query("device_id"),
			Platform:   c.Query("platform"),
			OSVersion:  c.Query("os_version"),
			AppVersion: c.Query("app_version"),
		}
	}

	// Override event type
	eventReq.EventType = string(models.SessionStart)
	eventReq.EventName = "session_start"

	// Record event
	tc.RecordEvent(c)
}

// EndSession ends a session for a device
// @Summary End session
// @Description End a session for a device
// @Tags telemetry
// @Accept json
// @Produce json
// @Param session body dtos.EventRequest true "Session details"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/telemetry/session/end [post]
func (tc *TelemetryController) EndSession(c *gin.Context) {
	// Session end is essentially an event with type "session_end"
	c.Request.Body = http.NoBody // Clear existing body to avoid binding issues
	c.Request.Header.Set("Content-Type", "application/json")

	var eventReq dtos.EventRequest
	if err := c.ShouldBindJSON(&eventReq); err != nil {
		eventReq = dtos.EventRequest{
			DeviceID:   c.Query("device_id"),
			Platform:   c.Query("platform"),
			OSVersion:  c.Query("os_version"),
			AppVersion: c.Query("app_version"),
		}
	}

	// Override event type
	eventReq.EventType = string(models.SessionEnd)
	eventReq.EventName = "session_end"

	// Record event
	tc.RecordEvent(c)
}

// RecordInstallation records an app installation
// @Summary Record installation
// @Description Record a new app installation
// @Tags telemetry
// @Accept json
// @Produce json
// @Param installation body dtos.InstallationRequest true "Installation details"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/telemetry/install [post]
func (tc *TelemetryController) RecordInstallation(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}

	var installReq dtos.InstallationRequest
	if err := c.ShouldBindJSON(&installReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find or create device
	device, err := tc.findOrCreateDevice(
		c,
		projectID.(uuid.UUID),
		installReq.DeviceID,
		installReq.Platform,
		installReq.OsVersion,
		installReq.AppVersion,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create or update device"})
		return
	}

	// Create installation event
	event := models.Event{
		ProjectID:  projectID.(uuid.UUID),
		DeviceID:   device.ID,
		EventType:  models.Install,
		EventName:  "install",
		Parameters: installReq.Properties,
		Timestamp:  time.Now(),
		ReceivedAt: time.Now(),
	}

	if err := tc.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record installation"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Installation recorded successfully"})
}
