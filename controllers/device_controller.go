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

// DeviceController handles device registration and management
type DeviceController struct {
	DB *gorm.DB
}

// NewDeviceController creates a new DeviceController instance
func NewDeviceController(db *gorm.DB) *DeviceController {
	return &DeviceController{DB: db}
}

// CreateDevice registers a new device or updates an existing one
// @Summary Register device
// @Description Register a new device or update an existing one
// @Tags devices
// @Accept json
// @Produce json
// @Param device body dtos.DeviceRequest true "Device details"
// @Security ApiKeyAuth
// @Success 200 {object} dtos.DeviceResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/devices/register [post]
func (dc *DeviceController) CreateDevice(c *gin.Context) {
	// Get project ID from context (set by ApiKeyMiddleware)
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Error: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Bind request body
	var request dtos.CreateDeviceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Error: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Try to find the device
	var device models.Device
	result := dc.DB.Where("project_id = ? AND machine_code = ?", projectID, request.Identifier).First(&device)

	// Client IP address
	ipAddress := c.ClientIP()

	// Device found - update it
	if result.Error == nil {
		device.LastSeen = time.Now()

		// Only update these if provided
		if request.AppVersion != "" {
			device.AppVersion = request.AppVersion
		}

		if request.PlatformVersion != "" {
			device.PlatformVersion = request.PlatformVersion
		}

		// Update IP and country if they've changed
		if device.IpAddress != ipAddress {
			device.IpAddress = ipAddress

			// Get updated country from IP address
			country, err := utils.GetCountryFromIP(ipAddress)
			if err == nil && country != "Unknown" {
				device.Country = country
			}
		}

		// Save the device
		if err := dc.DB.Save(&device).Error; err != nil {
			response := dtos.ErrorResponse{
				Error: "Failed to update device",
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}

		// Create response DTO
		response := dtos.CreateDeviceResponse{
			ID: device.ID,
		}

		// Return response
		c.JSON(http.StatusOK, response)
		return
	}

	// Device doesn't exist, create a new one
	// Get country from IP address
	country, err := utils.GetCountryFromIP(ipAddress)
	if err != nil {
		// If geolocation fails, continue with unknown country
		country = "Unknown"
	}

	// Create new device
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

	// Create new device
	if err := dc.DB.Create(&newDevice).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to create device",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.CreateDeviceResponse{
		ID: newDevice.ID,
	}

	// Return response
	c.JSON(http.StatusCreated, response)
}

// GetDevice retrieves a device by ID
// @Summary Get device
// @Description Get device by ID
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Security ApiKeyAuth
// @Success 200 {object} dtos.DeviceResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/devices/{id} [get]
func (dc *DeviceController) GetDevice(c *gin.Context) {
	// Get project ID from context (set by ApiKeyMiddleware)
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Error: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get device ID from URL
	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response := dtos.ErrorResponse{
			Error: "Invalid device ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get device
	var device models.Device
	if err := dc.DB.Where("id = ? AND project_id = ?", deviceID, projectID).First(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Device not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Create response DTO
	response := dtos.GetDeviceResponse{
		ID:              device.ID,
		Identifier:      device.Identifier,
		Platform:        device.Platform,
		PlatformVersion: device.PlatformVersion,
		AppVersion:      device.AppVersion,
		FirstSeen:       device.FirstSeen,
		LastSeen:        device.LastSeen,
	}

	c.JSON(http.StatusOK, response)
}

// GetDevices retrieves a list of devices with filtering options
// @Summary Get devices
// @Description Get a list of devices with filtering options
// @Tags devices
// @Accept json
// @Produce json
// @Param platform query string false "Filter by platform"
// @Param start_date query string false "Filter by first seen date (ISO 8601)"
// @Param end_date query string false "Filter by last seen date (ISO 8601)"
// @Param limit query int false "Result limit (default 20)"
// @Param offset query int false "Result offset (default 0)"
// @Security BearerAuth
// @Success 200 {object} dtos.GetDevicesResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/devices [get]
func (dc *DeviceController) GetDevices(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Error: "Not authenticated",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Bind query parameters
	var query dtos.GetDevicesRequestQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response := dtos.ErrorResponse{
			Error: "Invalid query parameters",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Set default pagination values if not provided
	if query.Limit <= 0 {
		query.Limit = 20
	} else if query.Limit > 100 {
		query.Limit = 100 // Cap at 100 for performance
	}

	if query.Offset < 0 {
		query.Offset = 0
	}

	// Start building the query - only include devices from projects owned by the user
	dbQuery := dc.DB.Model(&models.Device{}).
		Joins("JOIN projects ON devices.project_id = projects.id").
		Where("projects.owner_id = ?", userID)

	// Apply filters
	if query.Platform != "" {
		dbQuery = dbQuery.Where("devices.platform = ?", query.Platform)
	}

	if query.StartDate != "" {
		dbQuery = dbQuery.Where("devices.first_seen >= ?", query.StartDate)
	}

	if query.EndDate != "" {
		dbQuery = dbQuery.Where("devices.last_seen <= ?", query.EndDate)
	}

	// Count total before pagination
	var totalCount int64
	if err := dbQuery.Count(&totalCount).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to count devices",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Apply pagination and get results
	var devices []models.Device
	if err := dbQuery.Order("devices.last_seen DESC").
		Limit(query.Limit).
		Offset(query.Offset).
		Find(&devices).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to retrieve devices",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Map to response DTOs
	deviceResponses := make([]dtos.GetDeviceResponse, len(devices))
	for i, device := range devices {
		deviceResponses[i] = dtos.GetDeviceResponse{
			ID:              device.ID,
			Identifier:      device.Identifier,
			Platform:        device.Platform,
			PlatformVersion: device.PlatformVersion,
			AppVersion:      device.AppVersion,
			FirstSeen:       device.FirstSeen,
			LastSeen:        device.LastSeen,
		}
	}

	// Create response
	response := dtos.GetDevicesResponse{
		Devices:    deviceResponses,
		TotalCount: totalCount,
		Limit:      query.Limit,
		Offset:     query.Offset,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// UpdateDevice updates a device by ID
// @Summary Update device
// @Description Update a device by ID
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Param device body dtos.UpdateDeviceRequest true "Device update details"
// @Security ApiKeyAuth
// @Success 200 {object} dtos.DeviceResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/devices/{id} [patch]
func (dc *DeviceController) UpdateDevice(c *gin.Context) {
	// Get project ID from context (set by ApiKeyMiddleware)
	projectID, exists := c.Get("project_id")
	if !exists {
		response := dtos.ErrorResponse{
			Error: "Project not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get device ID from URL
	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response := dtos.ErrorResponse{
			Error: "Invalid device ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Bind request body
	var request dtos.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Error: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get device to update
	var device models.Device
	if err := dc.DB.Where("id = ? AND project_id = ?", deviceID, projectID).First(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Device not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Update fields if provided
	if request.AppVersion != "" {
		device.AppVersion = request.AppVersion
	}

	if request.PlatformVersion != "" {
		device.PlatformVersion = request.PlatformVersion
	}

	// Always update last seen timestamp
	device.LastSeen = time.Now()

	// Save changes
	if err := dc.DB.Save(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to update device",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.GetDeviceResponse{
		ID:              device.ID,
		Identifier:      device.Identifier,
		Platform:        device.Platform,
		PlatformVersion: device.PlatformVersion,
		AppVersion:      device.AppVersion,
		FirstSeen:       device.FirstSeen,
		LastSeen:        device.LastSeen,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// DeleteDevice deletes a device
// @Summary Delete device
// @Description Delete a device
// @Tags devices
// @Accept json
// @Produce json
// @Param id path string true "Device ID"
// @Security BearerAuth
// @Success 200 {object} dtos.DeleteDeviceResponse
// @Failure 404 {object} map[string]string
// @Router /api/v1/devices/{id} [delete]
func (dc *DeviceController) DeleteDevice(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Error: "Not authenticated",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get device ID from URL
	deviceID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response := dtos.ErrorResponse{
			Error: "Invalid device ID",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Get the device to check ownership
	var device models.Device
	if err := dc.DB.Preload("Project").First(&device, "id = ?", deviceID).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Device not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Check if user owns the project
	if device.Project.OwnerID != userID.(uuid.UUID) {
		response := dtos.ErrorResponse{
			Error: "Access denied",
		}
		c.JSON(http.StatusForbidden, response)
		return
	}

	// Delete device (this will be a soft delete due to gorm settings)
	if err := dc.DB.Delete(&device).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to delete device",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.DeleteDeviceResponse{
		Message: "Device deleted successfully",
	}

	// Return response
	c.JSON(http.StatusOK, response)
}
