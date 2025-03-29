package controllers

import (
	"net/http"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthController handles authentication-related endpoints
type AuthController struct {
	DB *gorm.DB
}

// NewAuthController creates a new AuthController instance
func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

// Login authenticates a user and returns a JWT token
// @Summary Login user
// @Description Authenticate user and receive JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body dtos.LoginRequest true "Login credentials"
// @Success 200 {object} dtos.LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	// Bind request body
	var loginReq dtos.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	// Find user by email
	var user models.User
	if err := ac.DB.Where("email = ?", loginReq.Email).First(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid credentials",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(loginReq.Password, user.Password) {
		response := dtos.ErrorResponse{
			Message: "Invalid credentials",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Create token - use new utility function instead of local helper
	token, expiresAt, err := utils.CreateToken(user)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to create token",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Save token to database
	authToken := models.AuthToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := ac.DB.Create(&authToken).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to create token",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	// Create response
	response := dtos.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	c.JSON(http.StatusOK, response)
}

// Me returns the current user information
// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.MeResponse
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/me [get]
func (ac *AuthController) Me(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Get user
	var user models.User
	if err := ac.DB.First(&user, "id = ?", userID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Create response using MeResponse DTO
	response := dtos.MeResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Return response
	c.JSON(http.StatusOK, response)
}

// Logout invalidates the current token
// @Summary Logout user
// @Description Invalidate current JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.LogoutResponse
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	// Get auth header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response := dtos.ErrorResponse{
			Message: "Authorization header is required",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	// Extract token
	tokenString := authHeader[7:] // Remove "Bearer " prefix

	// Delete token from database
	ac.DB.Where("token = ?", tokenString).Delete(&models.AuthToken{})

	// Create response using LogoutResponse DTO
	response := dtos.LogoutResponse{
		Message: "Logged out successfully",
	}

	c.JSON(http.StatusOK, response)
}
