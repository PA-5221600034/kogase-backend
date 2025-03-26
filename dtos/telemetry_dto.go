package dtos

import (
	"time"
)

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

// InstallationRequest represents an installation request
type InstallationRequest struct {
	DeviceID   string         `json:"device_id" binding:"required"`
	Platform   string         `json:"platform" binding:"required"`
	AppVersion string         `json:"app_version" binding:"required"`
	OsVersion  string         `json:"os_version" binding:"required"`
	Properties map[string]any `json:"properties"`
}
