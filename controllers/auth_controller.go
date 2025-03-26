package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/atqamz/kogase-backend/middleware"
	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/utils"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AuthController handles authentication-related endpoints
type AuthController struct {
	DB *gorm.DB
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      struct {
		ID    uuid.UUID `json:"id"`
		Email string    `json:"email"`
		Name  string    `json:"name"`
		Role  string    `json:"role"`
	} `json:"user"`
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
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /auth/login [post]
func (ac *AuthController) Login(c *gin.Context) {
	var loginReq LoginRequest
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
	var response LoginResponse
	response.Token = token
	response.ExpiresAt = expiresAt
	response.User.ID = user.ID
	response.User.Email = user.Email
	response.User.Name = user.Name
	response.User.Role = user.Role

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
// @Router /auth/me [get]
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
// @Router /auth/logout [post]
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

// Register registers a new user
// @Summary Register new user
// @Description Register a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body RegisterRequest true "Registration details"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 409 {object} map[string]string
// @Router /auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var registerReq RegisterRequest
	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// Check if email already exists
	var existingUser models.User
	if err := ac.DB.Where("email = ?", registerReq.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email already in use"})
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(registerReq.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user
	user := models.User{
		Email:    registerReq.Email,
		Password: hashedPassword,
		Name:     registerReq.Name,
		Role:     "developer", // Default role
	}

	if err := ac.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Hide password in response
	user.Password = ""

	c.JSON(http.StatusCreated, user)
}

// RegisterRequest represents the registration payload
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
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
		Role:   user.Role,
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
