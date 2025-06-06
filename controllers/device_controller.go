package controllers

import (
	"net/http"
	"time"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DeviceController struct {
	DB *gorm.DB
}

func NewDeviceController(db *gorm.DB) *DeviceController {
	return &DeviceController{DB: db}
}

// CreateOrUpdateDevice godoc
// @Summary Create or update device
// @Description Create a new device or update an existing one
// @Tags devices
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param device body dtos.CreateOrUpdateDeviceRequest true "Device details"
// @Success 200 {object} dtos.CreateOrUpdateDeviceResponse
// @Success 201 {object} dtos.CreateOrUpdateDeviceResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /devices [post]
func (dc *DeviceController) CreateOrUpdateDevice(c *gin.Context) {
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var request dtos.CreateOrUpdateDeviceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var device models.Device
	result := dc.DB.Model(&models.Device{}).
		Where("project_id = ? AND identifier = ?", projectID, request.Identifier).
		First(&device)

	ipAddress := c.ClientIP()

	if result.Error == nil {
		device.LastSeen = time.Now()

		if request.AppVersion != "" {
			device.AppVersion = request.AppVersion
		}

		if request.PlatformVersion != "" {
			device.PlatformVersion = request.PlatformVersion
		}

		if device.IpAddress != ipAddress {
			device.IpAddress = ipAddress

			country, err := utils.GetCountryFromIP(ipAddress)
			if err == nil && country != "Unknown" {
				device.Country = country
			}
		}

		if err := dc.DB.Save(&device).Error; err != nil {
			response := dtos.ErrorResponse{
				Message: "Failed to update device",
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		resultResponse := dtos.CreateOrUpdateDeviceResponse{
			DeviceID:        device.ID.String(),
			Identifier:      device.Identifier,
			Platform:        device.Platform,
			PlatformVersion: device.PlatformVersion,
			AppVersion:      device.AppVersion,
			FirstSeen:       device.FirstSeen,
			LastSeen:        device.LastSeen,
			IpAddress:       device.IpAddress,
			Country:         device.Country,
		}

		c.JSON(http.StatusOK, resultResponse)
		return
	}

	country, err := utils.GetCountryFromIP(ipAddress)
	if err != nil {
		country = "Unknown"
	}

	newDevice := models.Device{
		ProjectID:       projectID.(uuid.UUID),
		Identifier:      request.Identifier,
		Platform:        request.Platform,
		PlatformVersion: request.PlatformVersion,
		AppVersion:      request.AppVersion,
		FirstSeen:       time.Now(),
		LastSeen:        time.Now(),
		IpAddress:       ipAddress,
		Country:         country,
	}
	if err := dc.DB.Create(&newDevice).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to create device",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	event := models.Event{
		ProjectID:  projectID.(uuid.UUID),
		DeviceID:   newDevice.ID,
		EventType:  "predefined",
		EventName:  "install",
		Timestamp:  time.Now(),
		ReceivedAt: time.Now(),
	}
	if err := dc.DB.Create(&event).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to record install event",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.CreateOrUpdateDeviceResponse{
		DeviceID:        newDevice.ID.String(),
		Identifier:      newDevice.Identifier,
		Platform:        newDevice.Platform,
		PlatformVersion: newDevice.PlatformVersion,
		AppVersion:      newDevice.AppVersion,
		FirstSeen:       newDevice.FirstSeen,
		LastSeen:        newDevice.LastSeen,
		IpAddress:       newDevice.IpAddress,
		Country:         newDevice.Country,
	}

	c.JSON(http.StatusCreated, resultResponse)
}

// GetDevices godoc
// @Summary Get all devices
// @Description Retrieve a list of all devices with pagination
// @Tags devices
// @Produce json
// @Security BearerAuth
// @Param platform query string false "Filter by platform"
// @Param limit query int false "Limit results (default 20, max 100)"
// @Param offset query int false "Offset results (default 0)"
// @Success 200 {object} dtos.GetDevicesResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /devices [get]
func (dc *DeviceController) GetDevices(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Not authenticated",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var query dtos.GetDevicesRequestQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid query parameters",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if query.Limit <= 0 {
		query.Limit = 20
	} else if query.Limit > 100 {
		query.Limit = 100
	}

	if query.Offset < 0 {
		query.Offset = 0
	}

	dbQuery := dc.DB.Model(&models.Device{})

	if query.Platform != "" {
		dbQuery = dbQuery.Where("devices.platform = ?", query.Platform)
	}

	var totalCount int64
	if err := dbQuery.Count(&totalCount).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to count devices",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	var devices []models.Device
	if err := dbQuery.Order("devices.last_seen DESC").
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&devices).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to retrieve devices",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	deviceResponses := make([]dtos.GetDeviceResponse, len(devices))
	for i, device := range devices {
		deviceResponses[i] = dtos.GetDeviceResponse{
			DeviceID:        device.ID.String(),
			Identifier:      device.Identifier,
			Platform:        device.Platform,
			PlatformVersion: device.PlatformVersion,
			AppVersion:      device.AppVersion,
			FirstSeen:       device.FirstSeen,
			LastSeen:        device.LastSeen,
			IpAddress:       device.IpAddress,
			Country:         device.Country,
		}
	}

	resultResponse := dtos.GetDevicesResponse{
		Devices:    deviceResponses,
		TotalCount: int(totalCount),
		Limit:      query.Limit,
		Offset:     query.Offset,
	}

	c.JSON(http.StatusOK, resultResponse)
}

// GetDevice godoc
// @Summary Get a device by ID
// @Description Retrieve a specific device by its ID
// @Tags devices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Success 200 {object} dtos.GetDeviceResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /devices/{id} [get]
func (dc *DeviceController) GetDevice(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid device ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var device models.Device
	if err := dc.DB.Model(&models.Device{}).
		Where("id = ?", deviceID).
		First(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Device not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	resultResponse := dtos.GetDeviceResponse{
		DeviceID:        device.ID.String(),
		Identifier:      device.Identifier,
		Platform:        device.Platform,
		PlatformVersion: device.PlatformVersion,
		AppVersion:      device.AppVersion,
		FirstSeen:       device.FirstSeen,
		LastSeen:        device.LastSeen,
		IpAddress:       device.IpAddress,
		Country:         device.Country,
	}

	c.JSON(http.StatusOK, resultResponse)
}

// DeleteDevice godoc
// @Summary Delete a device
// @Description Delete a device by its ID
// @Tags devices
// @Produce json
// @Security BearerAuth
// @Param id path string true "Device ID"
// @Success 200 {object} dtos.DeleteDeviceResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /devices/{id} [delete]
func (dc *DeviceController) DeleteDevice(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "Not authenticated",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid device ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var device models.Device
	if err := dc.DB.Preload("Project").First(&device, "id = ?", deviceID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Device not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	if err := dc.DB.Delete(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to delete device",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.DeleteDeviceResponse{
		Message: "Device deleted successfully",
	}

	c.JSON(http.StatusOK, resultResponse)
}
