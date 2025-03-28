package dtos

import (
	"time"

	"github.com/google/uuid"
)

// CreateDeviceRequest represents a request to register or update a device
type CreateDeviceRequest struct {
	Identifier      string `json:"identifier" binding:"required"`
	Platform        string `json:"platform" binding:"required"`
	PlatformVersion string `json:"platform_version" binding:"required"`
	AppVersion      string `json:"app_version" binding:"required"`
}

// CreateDeviceResponse represents a response containing device information
type CreateDeviceResponse struct {
	ID uuid.UUID `json:"id"`
}

// GetDeviceResponse represents a response containing device information
type GetDeviceResponse struct {
	ID              uuid.UUID `json:"id"`
	Identifier      string    `json:"identifier"`
	Platform        string    `json:"platform"`
	PlatformVersion string    `json:"platform_version"`
	AppVersion      string    `json:"app_version"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	IpAddress       string    `json:"ip_address,omitempty"`
	Country         string    `json:"country,omitempty"`
}

// GetDevicesRequestQuery represents a request query to list devices with filters
type GetDevicesRequestQuery struct {
	Platform  string `form:"platform" json:"platform,omitempty"`
	StartDate string `form:"start_date" json:"start_date,omitempty"`
	EndDate   string `form:"end_date" json:"end_date,omitempty"`
	Limit     int    `form:"limit,default=20" json:"limit,omitempty"`
	Offset    int    `form:"offset,default=0" json:"offset,omitempty"`
}

// GetDevicesResponse represents a paginated list of devices
type GetDevicesResponse struct {
	Devices    []GetDeviceResponse `json:"devices"`
	TotalCount int64               `json:"total_count"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
}

// UpdateDeviceRequest represents a request to update a device
type UpdateDeviceRequest struct {
	Identifier      string `json:"identifier,omitempty"`
	Platform        string `json:"platform,omitempty"`
	PlatformVersion string `json:"platform_version,omitempty"`
	AppVersion      string `json:"app_version,omitempty"`
}

// UpdateDeviceResponse represent a response for a device update
type UpdateDeviceResponse struct {
	ID              uuid.UUID `json:"id"`
	Identifier      string    `json:"identifier"`
	Platform        string    `json:"platform"`
	PlatformVersion string    `json:"platform_version"`
	AppVersion      string    `json:"app_version"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	IpAddress       string    `json:"ip_address,omitempty"`
	Country         string    `json:"country,omitempty"`
}

// DeleteDeviceResponse represents a response for a device deletion
type DeleteDeviceResponse struct {
	Message string `json:"message"`
}
