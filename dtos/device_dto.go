package dtos

import (
	"time"

	"github.com/atqamz/kogase-backend/models"
)

type CreateOrUpdateDeviceRequest struct {
	Identifier      string `json:"identifier" binding:"required"`
	Platform        string `json:"platform" binding:"required"`
	PlatformVersion string `json:"platform_version" binding:"required"`
	AppVersion      string `json:"app_version" binding:"required"`
}

type CreateOrUpdateDeviceResponse struct {
	DeviceID        string    `json:"device_id"`
	Identifier      string    `json:"identifier"`
	Platform        string    `json:"platform"`
	PlatformVersion string    `json:"platform_version"`
	AppVersion      string    `json:"app_version"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	IpAddress       string    `json:"ip_address"`
	Country         string    `json:"country"`
}

type GetDevicesRequestQuery struct {
	Platform string `form:"platform" json:"platform,omitempty"`
	Limit    int    `form:"limit,default=20" json:"limit,omitempty"`
	Offset   int    `form:"offset,default=0" json:"offset,omitempty"`
}

type GetDevicesResponse struct {
	Devices    []GetDeviceResponse `json:"devices"`
	TotalCount int                 `json:"total_count"`
	Limit      int                 `json:"limit"`
	Offset     int                 `json:"offset"`
}

type GetDeviceResponse struct {
	DeviceID        string    `json:"device_id"`
	Identifier      string    `json:"identifier"`
	Platform        string    `json:"platform"`
	PlatformVersion string    `json:"platform_version"`
	AppVersion      string    `json:"app_version"`
	FirstSeen       time.Time `json:"first_seen"`
	LastSeen        time.Time `json:"last_seen"`
	IpAddress       string    `json:"ip_address"`
	Country         string    `json:"country"`
}

type GetDeviceResponseDetail struct {
	GetDeviceResponse
	Events []models.Event `json:"events"`
}

type DeleteDeviceResponse struct {
	Message string `json:"message"`
}
