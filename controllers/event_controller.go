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

type EventController struct {
	DB *gorm.DB
}

func NewEventController(db *gorm.DB) *EventController {
	return &EventController{DB: db}
}

// RecordEvent godoc
// @Summary Record a single event
// @Description Record a new telemetry event from a device
// @Tags events
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param event body dtos.RecordEventRequest true "Event details"
// @Success 201 {object} dtos.RecordEventResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /events [post]
func (tc *EventController) RecordEvent(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var request dtos.RecordEventRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var device models.Device
	if err := tc.DB.Where("identifier = ? AND project_id = ?", request.Identifier, projectID).First(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Device not found or doesn't belong to this project",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	device.LastSeen = time.Now()
	if err := tc.DB.Save(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to update device last seen",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

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

	resultResponse := dtos.RecordEventResponse{
		Message: "Event recorded successfully",
	}

	c.JSON(http.StatusCreated, resultResponse)
}

// RecordEvents godoc
// @Summary Record multiple events
// @Description Record a batch of telemetry events from a device
// @Tags events
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param events body dtos.RecordEventsRequest true "Batch of events"
// @Success 201 {object} dtos.RecordEventsResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /events/batch [post]
func (tc *EventController) RecordEvents(c *gin.Context) {
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

	err := tc.DB.Transaction(func(tx *gorm.DB) error {
		devices := make(map[string]models.Device)

		for _, eventReq := range request.Events {
			if device, exists := devices[eventReq.Identifier]; !exists {
				if err := tx.Where("project_id = ? AND identifier = ?", projectID, eventReq.Identifier).First(&device).Error; err != nil {
					return err
				}

				device.LastSeen = time.Now()
				if err := tx.Save(&device).Error; err != nil {
					return err
				}

				devices[eventReq.Identifier] = device
			}

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

	resultResponse := dtos.RecordEventsResponse{
		Message: "Events recorded successfully",
		Count:   len(request.Events),
	}

	c.JSON(http.StatusCreated, resultResponse)
}

// GetEvents godoc
// @Summary Get events
// @Description Retrieve events with filtering and pagination
// @Tags events
// @Produce json
// @Security BearerAuth
// @Param project_id query string false "Filter by project ID"
// @Param event_type query string false "Filter by event type"
// @Param event_name query string false "Filter by event name"
// @Param from_date query string false "Filter by start date (RFC3339)"
// @Param to_date query string false "Filter by end date (RFC3339)"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} dtos.GetEventsResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /events [get]
func (tc *EventController) GetEvents(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var request dtos.GetEventsRequestQuery
	if err := c.ShouldBindQuery(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	dbQuery := tc.DB.Model(&models.Event{})
	if request.ProjectID != "" {
		dbQuery = dbQuery.Where("project_id = ?", request.ProjectID)
	}
	if request.FromDate != "" {
		dbQuery = dbQuery.Where("timestamp >= ?", request.FromDate)
	}
	if request.ToDate != "" {
		dbQuery = dbQuery.Where("timestamp <= ?", request.ToDate)
	}
	if request.EventType != "" {
		dbQuery = dbQuery.Where("event_type = ?", request.EventType)
	}
	if request.EventName != "" {
		dbQuery = dbQuery.Where("event_name = ?", request.EventName)
	}

	var totalCount int64
	if err := dbQuery.Count(&totalCount).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to count events",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	var events []models.Event
	if err := dbQuery.Order("timestamp DESC").Limit(request.Limit).Offset(request.Offset).Find(&events).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to get events",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	eventsResponse := make([]dtos.GetEventResponse, len(events))
	for i, event := range events {
		eventsResponse[i] = dtos.GetEventResponse{
			EventID:    event.ID.String(),
			EventType:  event.EventType,
			EventName:  event.EventName,
			Payloads:   event.Payloads,
			Timestamp:  event.Timestamp.Format(time.RFC3339),
			ReceivedAt: event.ReceivedAt.Format(time.RFC3339),
		}
	}

	resultResponse := dtos.GetEventsResponse{
		Events: eventsResponse,
		Total:  int(totalCount),
	}

	c.JSON(http.StatusOK, resultResponse)
}

// GetEvent godoc
// @Summary Get event by ID
// @Description Retrieve a specific event by its ID
// @Tags events
// @Produce json
// @Security BearerAuth
// @Param id path string true "Event ID"
// @Success 200 {object} dtos.GetEventResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /events/{id} [get]
func (tc *EventController) GetEvent(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var request dtos.GetEventRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var event models.Event
	if err := tc.DB.Where("id = ?", request.EventID).First(&event).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Event not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	resultResponse := dtos.GetEventResponse{
		EventID:    event.ID.String(),
		EventType:  event.EventType,
		EventName:  event.EventName,
		Payloads:   event.Payloads,
		Timestamp:  event.Timestamp.Format(time.RFC3339),
		ReceivedAt: event.ReceivedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, resultResponse)
}
