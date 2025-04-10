package controllers

import (
	"net/http"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthController struct {
	DB *gorm.DB
}

func NewAuthController(db *gorm.DB) *AuthController {
	return &AuthController{DB: db}
}

func (ac *AuthController) Login(c *gin.Context) {
	var request dtos.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var user models.User
	if err := ac.DB.Model(&models.User{}).
		Where("email = ?", request.Email).
		First(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid credentials",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	if !utils.CheckPasswordHash(request.Password, user.Password) {
		response := dtos.ErrorResponse{
			Message: "Invalid credentials",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	token, expiresAt, err := utils.CreateToken(user)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to create token",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

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

	resultResponse := dtos.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
	}

	c.JSON(http.StatusOK, resultResponse)
}

func (ac *AuthController) Me(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var user models.User
	if err := ac.DB.First(&user, "id = ?", userID).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	resultResponse := dtos.MeResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	c.JSON(http.StatusOK, resultResponse)
}

func (ac *AuthController) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		response := dtos.ErrorResponse{
			Message: "Authorization header is required",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	tokenString := authHeader[7:]

	ac.DB.Model(&models.AuthToken{}).
		Where("token = ?", tokenString).
		Delete(&models.AuthToken{})

	resultResponse := dtos.LogoutResponse{
		Message: "Logged out successfully",
	}

	c.JSON(http.StatusOK, resultResponse)
}
