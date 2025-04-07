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

// EventController handles telemetry-related endpoints
type EventController struct {
	DB *gorm.DB
}

// NewEventController creates a new EventController instance
func NewEventController(db *gorm.DB) *EventController {
	return &EventController{DB: db}
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
func (tc *EventController) RecordEvent(c *gin.Context) {
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
	if err := tc.DB.Where("identifier = ? AND project_id = ?", request.Identifier, projectID).First(&device).Error; err != nil {
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
		DeviceID:   device.ID,
		EventType:  request.EventType,
		EventName:  request.EventName,
		Payloads:   request.Payloads,
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
func (tc *EventController) RecordEvents(c *gin.Context) {
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
		devices := make(map[string]models.Device)

		for _, eventReq := range request.Events {
			// Verify device exists and belongs to this project
			if device, exists := devices[eventReq.Identifier]; !exists {
				if err := tx.Where("project_id = ? AND identifier = ?", projectID, eventReq.Identifier).First(&device).Error; err != nil {
					return err
				}

				// Update device last seen
				device.LastSeen = time.Now()
				if err := tx.Save(&device).Error; err != nil {
					return err
				}

				// Mark device as verified
				devices[eventReq.Identifier] = device
			}

			// Create event
			timestamp := time.Now()
			if eventReq.Timestamp != nil {
				timestamp = *eventReq.Timestamp
			}
			event := models.Event{
				ProjectID:  projectID.(uuid.UUID),
				DeviceID:   devices[eventReq.Identifier].ID,
				EventType:  eventReq.EventType,
				EventName:  eventReq.EventName,
				Payloads:   eventReq.Payloads,
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
