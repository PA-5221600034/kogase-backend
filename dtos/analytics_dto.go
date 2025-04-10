package dtos

import (
	"time"
)

type GetAnalyticsRequestQuery struct {
	ProjectID string    `form:"project_id" json:"project_id,omitempty"`
	FromDate  time.Time `form:"from_date" json:"from_date,omitempty"`
	ToDate    time.Time `form:"to_date" json:"to_date,omitempty"`
}

type GetAnalyticsResponse struct {
	DAU           int           `json:"dau" binding:"required"`
	MAU           int           `json:"mau" binding:"required"`
	TotalDuration time.Duration `json:"total_duration" binding:"required"`
	TotalInstalls int           `json:"total_installs" binding:"required"`
}
