package dtos

import (
	"time"
)

type RecordEventRequest struct {
	Identifier string                 `json:"identifier" binding:"required"`
	EventType  string                 `json:"event_type" binding:"required"`
	EventName  string                 `json:"event_name" binding:"required"`
	Payloads   map[string]interface{} `json:"payloads"`
	Timestamp  *time.Time             `json:"timestamp"`
}

type RecordEventResponse struct {
	Message string `json:"message"`
}

type RecordEventsRequest struct {
	Events []RecordEventRequest `json:"events" binding:"required"`
}

type RecordEventsResponse struct {
	Message string `json:"message"`
	Count   int    `json:"count"`
}

type GetEventsRequestQuery struct {
	ProjectID string `form:"project_id" json:"project_id,omitempty"`
	StartDate string `form:"start_date" json:"start_date,omitempty"`
	EndDate   string `form:"end_date" json:"end_date,omitempty"`
	EventType string `form:"event_type" json:"event_type,omitempty"`
	EventName string `form:"event_name" json:"event_name,omitempty"`
	Limit     int    `form:"limit" json:"limit,omitempty"`
	Offset    int    `form:"offset" json:"offset,omitempty"`
}

type GetEventsResponse struct {
	Events []GetEventResponse `json:"events"`
	Total  int                `json:"total"`
}

type GetEventRequest struct {
	EventID string `form:"event_id" json:"event_id,omitempty"`
}

type GetEventResponse struct {
	EventID    string                 `json:"event_id"`
	EventType  string                 `json:"event_type"`
	EventName  string                 `json:"event_name"`
	Payloads   map[string]interface{} `json:"payloads"`
	Timestamp  time.Time              `json:"timestamp"`
	ReceivedAt time.Time              `json:"received_at"`
}
