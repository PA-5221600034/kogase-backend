package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/middleware"
	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
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
	var loginReq dtos.LoginRequest
	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Find user by email
	var user models.User
	if err := ac.DB.Where("email = ?", loginReq.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Verify password
	if !utils.CheckPasswordHash(loginReq.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Create token
	token, expiresAt, err := createToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// Save token to database
	authToken := models.AuthToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := ac.DB.Create(&authToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	// Create response
	var response dtos.LoginResponse
	response.Token = token
	response.ExpiresAt = expiresAt
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.Name = user.Name

	c.JSON(http.StatusOK, response)
}

// Me returns the current user information
// @Summary Get current user
// @Description Get current authenticated user information
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.User
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/me [get]
func (ac *AuthController) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	var user models.User
	if err := ac.DB.First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// Logout invalidates the current token
// @Summary Logout user
// @Description Invalidate current JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /api/v1/auth/logout [post]
func (ac *AuthController) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
		return
	}

	// Extract token
	tokenString := authHeader[7:] // Remove "Bearer " prefix

	// Delete token from database
	ac.DB.Where("token = ?", tokenString).Delete(&models.AuthToken{})

	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// Helper function to create a JWT token
func createToken(user models.User) (string, time.Time, error) {
	// Get JWT expiry from env
	expiryStr := os.Getenv("JWT_EXPIRES_IN")
	if expiryStr == "" {
		expiryStr = "24h" // Default to 24 hours
	}

	// Parse the duration
	expiryDuration, err := time.ParseDuration(expiryStr)
	if err != nil {
		expiryDuration = 24 * time.Hour // Default to 24 hours
	}

	expiresAt := time.Now().Add(expiryDuration)

	// Create claims
	claims := middleware.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "kogase-api",
			Subject:   user.ID.String(),
		},
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token
	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}
