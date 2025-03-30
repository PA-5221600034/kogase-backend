package dtos

import (
	"time"

	"github.com/google/uuid"
)

type BeginSessionRequest struct {
	DeviceID uuid.UUID `json:"device_id" binding:"required"`
}

type BeginSessionResponse struct {
	Message string `json:"message"`
}

type FinishSessionRequest struct {
	DeviceID uuid.UUID `json:"device_id" binding:"required"`
}

type FinishSessionResponse struct {
	Message string `json:"message"`
}

type GetDeviceSessionsRequestQuery struct {
	ProjectID uuid.UUID `form:"project_id" json:"project_id,omitempty"`
	DeviceID  uuid.UUID `form:"device_id" json:"device_id,omitempty"`
	StartDate time.Time `form:"start_date" json:"start_date,omitempty"`
	EndDate   time.Time `form:"end_date" json:"end_date,omitempty"`
	Limit     int       `form:"limit,default=20" json:"limit,omitempty"`
	Offset    int       `form:"offset,default=0" json:"offset,omitempty"`
}

type GetProjectSessionsRequestQuery struct {
	ProjectID uuid.UUID `form:"project_id" json:"project_id,omitempty"`
	StartDate time.Time `form:"start_date" json:"start_date,omitempty"`
	EndDate   time.Time `form:"end_date" json:"end_date,omitempty"`
	Limit     int       `form:"limit,default=20" json:"limit,omitempty"`
	Offset    int       `form:"offset,default=0" json:"offset,omitempty"`
}

type GetAllSessionsRequestQuery struct {
	StartDate time.Time `form:"start_date" json:"start_date,omitempty"`
	EndDate   time.Time `form:"end_date" json:"end_date,omitempty"`
	Limit     int       `form:"limit,default=20" json:"limit,omitempty"`
	Offset    int       `form:"offset,default=0" json:"offset,omitempty"`
}

type GetSessionResponse struct {
	ID        uuid.UUID `json:"id"`
	ProjectID uuid.UUID `json:"project_id"`
	DeviceID  uuid.UUID `json:"device_id"`
	BeginAt   time.Time `json:"begin_at"`
	EndAt     time.Time `json:"end_at"`
}

type GetSessionsResponse struct {
	Sessions []GetSessionResponse `json:"sessions"`
	Total    int64                `json:"total"`
	Limit    int                  `json:"limit"`
	Offset   int                  `json:"offset"`
}
