package dtos

import (
	"time"
)

// MetricsQuery represents a query for analytics metrics
type MetricsQuery struct {
	MetricType string     `form:"metric_type"`
	StartDate  *time.Time `form:"start_date"`
	EndDate    *time.Time `form:"end_date"`
	Period     string     `form:"period" binding:"omitempty,oneof=hourly daily weekly monthly yearly total"`
	Dimensions []string   `form:"dimensions"`
}

// EventsQuery represents a query for events
type EventsQuery struct {
	EventType string     `form:"event_type"`
	EventName string     `form:"event_name"`
	StartDate *time.Time `form:"start_date"`
	EndDate   *time.Time `form:"end_date"`
	DeviceID  string     `form:"device_id"`
	Platform  string     `form:"platform"`
	Limit     int        `form:"limit,default=100"`
	Offset    int        `form:"offset,default=0"`
}
