package dtos

import (
	"time"

	"github.com/atqamz/kogase-backend/models"
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

type EventQueryParams struct {
	ProjectID string `json:"project_id" binding:"required"`
	StartDate string `json:"start_date" binding:"required"`
	EndDate   string `json:"end_date" binding:"required"`
	EventType string `json:"event_type"`
	EventName string `json:"event_name"`
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
}

type EventQueryResponse struct {
	Events []models.Event `json:"events"`
	Total  int            `json:"total"`
}

type EventCountByTypeResponse struct {
	EventType string `json:"event_type"`
	Count     int    `json:"count"`
}

type EventCountByDayResponse struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}
