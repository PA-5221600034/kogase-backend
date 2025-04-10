package dtos

import (
	"time"
)

type BeginSessionRequest struct {
	Identifier string `json:"identifier" binding:"required"`
}

type BeginSessionResponse struct {
	SessionID string `json:"session_id"`
}

type EndSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

type EndSessionResponse struct {
	Message string `json:"message"`
}

type GetSessionsRequestQuery struct {
	ProjectID string    `form:"project_id" json:"project_id,omitempty"`
	FromDate  time.Time `form:"from_date" json:"from_date,omitempty"`
	ToDate    time.Time `form:"to_date" json:"to_date,omitempty"`
	Limit     int       `form:"limit,default=20" json:"limit,omitempty"`
	Offset    int       `form:"offset,default=0" json:"offset,omitempty"`
}

type GetSessionsResponse struct {
	Sessions []GetSessionResponse `json:"sessions"`
	Total    int64                `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
}

type GetSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

type GetSessionResponse struct {
	SessionID string        `json:"session_id"`
	BeginAt   time.Time     `json:"begin_at"`
	EndAt     time.Time     `json:"end_at"`
	Duration  time.Duration `json:"duration"`
}
