package controllers

import (
	"net/http"
	"time"

	"github.com/atqamz/kogase-backend/models"
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

// EventRequest represents a telemetry event request
type EventRequest struct {
	DeviceID   string                 `json:"device_id" binding:"required"`
	EventType  string                 `json:"event_type" binding:"required"`
	EventName  string                 `json:"event_name" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Timestamp  *time.Time             `json:"timestamp"`
	Platform   string                 `json:"platform" binding:"required"`
	OSVersion  string                 `json:"os_version" binding:"required"`
	AppVersion string                 `json:"app_version" binding:"required"`
}

// EventsRequest represents a batch of telemetry events
type EventsRequest struct {
	Events []EventRequest `json:"events" binding:"required"`
}

// RecordEvent records a single telemetry event
// @Summary Record event
// @Description Record a telemetry event
// @Tags telemetry
// @Accept json
// @Produce json
// @Param event body EventRequest true "Event details"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /telemetry/events [post]
func (tc *TelemetryController) RecordEvent(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}

	var eventReq EventRequest
	if err := c.ShouldBindJSON(&eventReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find or create device
	var device models.Device
	deviceResult := tc.DB.Where("project_id = ? AND device_id = ?", projectID, eventReq.DeviceID).First(&device)

	if deviceResult.Error != nil {
		// Device doesn't exist, create a new one
		device = models.Device{
			ProjectID:  projectID.(uuid.UUID),
			DeviceID:   eventReq.DeviceID,
			Platform:   eventReq.Platform,
			OSVersion:  eventReq.OSVersion,
			AppVersion: eventReq.AppVersion,
			FirstSeen:  time.Now(),
			LastSeen:   time.Now(),
			IPAddress:  c.ClientIP(),
			Country:    "", // TODO: Implement IP-based geolocation
		}

		if err := tc.DB.Create(&device).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create device"})
			return
		}
	} else {
		// Update device
		device.LastSeen = time.Now()
		device.AppVersion = eventReq.AppVersion
		device.OSVersion = eventReq.OSVersion
		device.IPAddress = c.ClientIP()
		tc.DB.Save(&device)
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
// @Param events body EventsRequest true "Events batch"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /telemetry/events/batch [post]
func (tc *TelemetryController) RecordEvents(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project not found"})
		return
	}

	var eventsReq EventsRequest
	if err := c.ShouldBindJSON(&eventsReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Process events in a transaction
	err := tc.DB.Transaction(func(tx *gorm.DB) error {
		// Keep track of devices to avoid multiple lookups
		devices := make(map[string]models.Device)

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
						IPAddress:  c.ClientIP(),
						Country:    "", // TODO: Implement IP-based geolocation
					}

					if err := tx.Create(&dbDevice).Error; err != nil {
						return err
					}
				} else {
					// Update device
					dbDevice.LastSeen = time.Now()
					dbDevice.AppVersion = eventReq.AppVersion
					dbDevice.OSVersion = eventReq.OSVersion
					dbDevice.IPAddress = c.ClientIP()
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
// @Param session body EventRequest true "Session start details"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /telemetry/session/start [post]
func (tc *TelemetryController) StartSession(c *gin.Context) {
	// Set event type to session_start
	c.Set("eventType", models.SessionStart)
	tc.RecordEvent(c)
}

// EndSession ends a session for a device
// @Summary End session
// @Description End a session for a device
// @Tags telemetry
// @Accept json
// @Produce json
// @Param session body EventRequest true "Session end details"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /telemetry/session/end [post]
func (tc *TelemetryController) EndSession(c *gin.Context) {
	// Set event type to session_end
	c.Set("eventType", models.SessionEnd)
	tc.RecordEvent(c)
}

// RecordInstall records an installation event
// @Summary Record install
// @Description Record an installation event
// @Tags telemetry
// @Accept json
// @Produce json
// @Param install body EventRequest true "Install details"
// @Security ApiKeyAuth
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /telemetry/install [post]
func (tc *TelemetryController) RecordInstall(c *gin.Context) {
	// Set event type to install
	c.Set("eventType", models.Install)
	tc.RecordEvent(c)
}

// RecordInstallation records a new installation
// @Summary Record installation
// @Description Record a new installation event
// @Tags sdk
// @Accept json
// @Produce json
// @Param installation body InstallationRequest true "Installation data"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /sdk/installation [post]
func (tc *TelemetryController) RecordInstallation(c *gin.Context) {
	projectID, _ := c.Get("project_id")

	var req InstallationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find or create device
	device := models.Device{
		ProjectID:  projectID.(uuid.UUID),
		DeviceID:   req.DeviceID,
		Platform:   req.Platform,
		AppVersion: req.AppVersion,
		OSVersion:  req.OsVersion,
		FirstSeen:  time.Now(),
		LastSeen:   time.Now(),
	}

	var existingDevice models.Device
	err := tc.DB.Where("project_id = ? AND device_id = ?", projectID, req.DeviceID).First(&existingDevice).Error
	if err != nil {
		if err := tc.DB.Create(&device).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create device"})
			return
		}
	} else {
		device.ID = existingDevice.ID
		device.FirstSeen = existingDevice.FirstSeen
		if err := tc.DB.Model(&device).Updates(map[string]interface{}{
			"platform":    req.Platform,
			"app_version": req.AppVersion,
			"os_version":  req.OsVersion,
			"last_seen":   time.Now(),
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update device"})
			return
		}
	}

	// Create installation event
	event := models.Event{
		ProjectID:  projectID.(uuid.UUID),
		DeviceID:   device.ID,
		EventType:  models.Install,
		EventName:  "installation",
		Parameters: req.Properties,
		Timestamp:  time.Now(),
		ReceivedAt: time.Now(),
	}

	if err := tc.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record installation event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// InstallationRequest represents an installation request
type InstallationRequest struct {
	DeviceID   string         `json:"device_id" binding:"required"`
	Platform   string         `json:"platform" binding:"required"`
	AppVersion string         `json:"app_version" binding:"required"`
	OsVersion  string         `json:"os_version" binding:"required"`
	Properties map[string]any `json:"properties"`
}
