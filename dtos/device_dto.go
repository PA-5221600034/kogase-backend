package dtos

import (
	"time"

	"github.com/atqamz/kogase-backend/models"
	"github.com/google/uuid"
)

type CreateOrUpdateDeviceRequest struct {
	Identifier      string `json:"identifier" binding:"required"`
	Platform        string `json:"platform" binding:"required"`
	PlatformVersion string `json:"platform_version" binding:"required"`
	AppVersion      string `json:"app_version" binding:"required"`
}

type GetDeviceResponse struct {
	DeviceID        uuid.UUID `json:"device_id"`
	Identifier      string    `json:"identifier"`
	Platform        string    `json:"platform"`
	PlatformVersion string    `json:"platform_version"`
	AppVersion      string    `json:"app_version"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	IpAddress       string    `json:"ip_address,omitempty"`
	Country         string    `json:"country,omitempty"`
}

type GetDeviceResponseDetail struct {
	DeviceID        uuid.UUID      `json:"device_id"`
	Identifier      string         `json:"identifier"`
	Platform        string         `json:"platform"`
	PlatformVersion string         `json:"platform_version"`
	AppVersion      string         `json:"app_version"`
	FirstSeen       time.Time      `json:"first_seen"`
	LastSeen        time.Time      `json:"last_seen"`
	IpAddress       string         `json:"ip_address,omitempty"`
	Country         string         `json:"country,omitempty"`
	Events          []models.Event `json:"events"`
}

type GetDevicesRequestQuery struct {
	Platform string `form:"platform" json:"platform,omitempty"`
	Limit    int    `form:"limit,default=20" json:"limit,omitempty"`
	Offset   int    `form:"offset,default=0" json:"offset,omitempty"`
}

type GetDevicesResponse struct {
	Devices    []GetDeviceResponse `json:"devices"`
	TotalCount int64               `json:"total_count"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
}

type UpdateDeviceRequest struct {
	Identifier      string `json:"identifier,omitempty"`
	Platform        string `json:"platform,omitempty"`
	PlatformVersion string `json:"platform_version,omitempty"`
	AppVersion      string `json:"app_version,omitempty"`
}

type CreateOrUpdateDeviceResponse struct {
	DeviceID        uuid.UUID `json:"device_id"`
	Identifier      string    `json:"identifier"`
	Platform        string    `json:"platform"`
	PlatformVersion string    `json:"platform_version"`
	AppVersion      string    `json:"app_version"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	IpAddress       string    `json:"ip_address,omitempty"`
	Country         string    `json:"country,omitempty"`
}

type DeleteDeviceResponse struct {
	Message string `json:"message"`
}
