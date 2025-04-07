package dtos

import (
	"time"
)

// RecordEventRequest represents a telemetry event request
type RecordEventRequest struct {
	Identifier string                 `json:"identifier" binding:"required"`
	EventType  string                 `json:"event_type" binding:"required"`
	EventName  string                 `json:"event_name" binding:"required"`
	Payloads   map[string]interface{} `json:"payloads"`
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
