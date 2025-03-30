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

// TelemetryController handles telemetry-related endpoints
type TelemetryController struct {
	DB *gorm.DB
}

// NewTelemetryController creates a new TelemetryController instance
func NewTelemetryController(db *gorm.DB) *TelemetryController {
	return &TelemetryController{DB: db}
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
	var request dtos.RecordEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Verify device exists and belongs to this project
	var device models.Device
	if err := tc.DB.Where("id = ? AND project_id = ?", request.DeviceID, projectID).First(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Device not found or doesn't belong to this project",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update device last seen
	device.LastSeen = time.Now()
	if err := tc.DB.Save(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to update device last seen",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create event
	timestamp := time.Now()
	if request.Timestamp != nil {
		timestamp = *request.Timestamp
	}
	event := models.Event{
		ProjectID:  projectID.(uuid.UUID),
		DeviceID:   request.DeviceID,
		EventType:  models.EventType(request.EventType),
		EventName:  request.EventName,
		Parameters: request.Parameters,
		Timestamp:  timestamp,
		ReceivedAt: time.Now(),
	}
	if err := tc.DB.Create(&event).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to record event",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.RecordEventResponse{
		Message: "Event recorded successfully",
	}

	// Return response
	c.JSON(http.StatusCreated, response)
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
	// Get project ID from context (set by ApiKeyMiddleware)
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var request dtos.RecordEventsRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Process events in a transaction
	err := tc.DB.Transaction(func(tx *gorm.DB) error {
		// Keep track of devices to avoid multiple lookups and updates
		devices := make(map[uuid.UUID]bool)

		for _, eventReq := range request.Events {
			// Verify device exists and belongs to this project
			if _, exists := devices[eventReq.DeviceID]; !exists {
				var device models.Device
				if err := tx.Where("id = ? AND project_id = ?", eventReq.DeviceID, projectID).First(&device).Error; err != nil {
					return err
				}

				// Update device last seen
				device.LastSeen = time.Now()
				if err := tx.Save(&device).Error; err != nil {
					return err
				}

				// Mark device as verified
				devices[eventReq.DeviceID] = true
			}

			// Create event
			timestamp := time.Now()
			if eventReq.Timestamp != nil {
				timestamp = *eventReq.Timestamp
			}
			event := models.Event{
				ProjectID:  projectID.(uuid.UUID),
				DeviceID:   eventReq.DeviceID,
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
		response := dtos.ErrorResponse{
			Message: "Failed to record events",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.RecordEventsResponse{
		Message: "Events recorded successfully",
		Count:   len(request.Events),
	}

	// Return response
	c.JSON(http.StatusCreated, response)
}

// RecordInstall records an app installation
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
func (tc *TelemetryController) RecordInstall(c *gin.Context) {
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
	var request dtos.RecordInstallRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Verify device exists and belongs to this project
	var device models.Device
	if err := tc.DB.Where("id = ? AND project_id = ?", request.DeviceID, projectID).First(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Device not found or doesn't belong to this project",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update device last seen
	device.LastSeen = time.Now()
	if err := tc.DB.Save(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to update device last seen",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create installation event
	event := models.Event{
		ProjectID:  projectID.(uuid.UUID),
		DeviceID:   request.DeviceID,
		EventType:  models.Install,
		EventName:  "install",
		Parameters: request.Parameters,
		Timestamp:  time.Now(),
		ReceivedAt: time.Now(),
	}

	if err := tc.DB.Create(&event).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to record installation",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.RecordInstallResponse{
		Message: "Installation recorded successfully",
	}

	// Return response
	c.JSON(http.StatusCreated, response)
}
