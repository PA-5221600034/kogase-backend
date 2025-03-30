package dtos

import (
	"time"

	"github.com/google/uuid"
)

// RecordEventRequest represents a telemetry event request
type RecordEventRequest struct {
	DeviceID   uuid.UUID              `json:"device_id" binding:"required"`
	EventType  string                 `json:"event_type" binding:"required"`
	EventName  string                 `json:"event_name" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Timestamp  *time.Time             `json:"timestamp"`
}

// RecordEventResponse represents a telemetry event response
type RecordEventResponse struct {
	Message string `json:"message"`
}

// RecordEventsRequest represents a batch of telemetry events
type RecordEventsRequest struct {
	Events []RecordEventRequest `json:"events" binding:"required"`
}

// RecordEventsResponse represents a batch of telemetry events response
type RecordEventsResponse struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

// RecordInstallRequest represents an installation request
type RecordInstallRequest struct {
	DeviceID   uuid.UUID              `json:"device_id" binding:"required"`
	Parameters map[string]interface{} `json:"parameters"`
	Timestamp  *time.Time             `json:"timestamp"`
}

// RecordInstallResponse represents an installation response
type RecordInstallResponse struct {
	Message string `json:"message"`
}
