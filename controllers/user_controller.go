package controllers

import (
	"net/http"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UserController handles user-related endpoints
type UserController struct {
	DB *gorm.DB
}

// NewUserController creates a new UserController instance
func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

// CreateUser creates a new user
// @Summary Create user
// @Description Create a new user
// @Tags users
// @Accept json
// @Produce json
// @Param user body dtos.RegisterRequest true "User details"
// @Security BearerAuth
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /api/v1/users [post]
func (uc *UserController) CreateUser(c *gin.Context) {
	// Parse request
	var userReq dtos.CreateUserRequest
	if err := c.ShouldBindJSON(&userReq); err != nil {
		response := dtos.ErrorResponse{
			Error: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := uc.DB.Where("email = ?", userReq.Email).First(&existingUser).Error; err == nil {
		response := dtos.ErrorResponse{
			Error: "Email already in use",
		}
		c.JSON(http.StatusConflict, response)
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(userReq.Password)
	if err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to hash password",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create user
	user := models.User{
		Email:    userReq.Email,
		Password: hashedPassword,
		Name:     userReq.Name,
	}
	if err := uc.DB.Create(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to create user",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Return user
	c.JSON(http.StatusCreated, user)
}

// GetUser returns a specific user by ID
// @Summary Get user
// @Description Get a specific user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users/{id} [get]
func (uc *UserController) GetUser(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exist := c.Get("user_id")
	if !exist {
		response := dtos.ErrorResponse{
			Error: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get user
	var user models.User
	if err := uc.DB.First(&user, "id = ?", userID).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "User not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Create response DTO
	response := dtos.GetUserResponseDetail{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}

	// Return user
	c.JSON(http.StatusOK, response)
}

// GetUsers returns all users
// @Summary List users
// @Description Get all users
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.User
// @Failure 401 {object} map[string]string
// @Router /api/v1/users [get]
func (uc *UserController) GetUsers(c *gin.Context) {
	// Get users
	var users []models.User
	if err := uc.DB.Find(&users).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to retrieve users",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.GetUsersResponse{
		Users: make([]dtos.GetUserResponse, len(users)),
	}
	for i, user := range users {
		response.Users[i] = dtos.GetUserResponse{
			ID:    user.ID,
			Email: user.Email,
			Name:  user.Name,
		}
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// UpdateUser updates a user
// @Summary Update user
// @Description Update user details
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Param user body dtos.UpdateUserRequest true "User details"
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users [patch]
func (uc *UserController) UpdateUser(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exist := c.Get("user_id")
	if !exist {
		response := dtos.ErrorResponse{
			Error: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get user
	var user models.User
	if err := uc.DB.First(&user, "id = ?", userID).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "User not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Parse request
	var updateReq dtos.UpdateUserRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response := dtos.ErrorResponse{
			Error: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Update only provided fields
	if updateReq.Name != "" {
		user.Name = updateReq.Name
	}

	// Update password if provided
	if updateReq.Password != "" {
		hashedPassword, err := utils.HashPassword(updateReq.Password)
		if err != nil {
			response := dtos.ErrorResponse{
				Error: "Failed to hash password",
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		user.Password = hashedPassword
	}

	// Save user
	if err := uc.DB.Save(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to update user",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.UpdateUserResponse{
		Email: user.Email,
		Name:  user.Name,
	}

	// Return user
	c.JSON(http.StatusOK, response)
}

// DeleteUser deletes a user
// @Summary Delete user
// @Description Delete a user
// @Tags users
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/users [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exist := c.Get("user_id")
	if !exist {
		response := dtos.ErrorResponse{
			Error: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get user
	var user models.User
	if err := uc.DB.First(&user, "id = ?", userID).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "User not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	// Delete user
	if err := uc.DB.Delete(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Error: "Failed to delete user",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response DTO
	response := dtos.DeleteUserResponse{
		Message: "User deleted successfully",
	}

	// Return response
	c.JSON(http.StatusOK, response)
}
