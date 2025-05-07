package controllers

import (
	"net/http"

	"github.com/atqamz/kogase-backend/dtos"
	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserController struct {
	DB *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Register a new user account
// @Tags users
// @Accept json
// @Produce json
// @Param user body dtos.CreateUserRequest true "User details"
// @Success 201 {object} dtos.CreateUserResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 409 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /users [post]
func (uc *UserController) CreateUser(c *gin.Context) {
	var userReq dtos.CreateUserRequest
	if err := c.ShouldBindJSON(&userReq); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	var existingUser models.User
	if err := uc.DB.Model(&models.User{}).
		Where("email = ?", userReq.Email).
		First(&existingUser).Error; err == nil {
		response := dtos.ErrorResponse{
			Message: "Email already in use",
		}
		c.JSON(http.StatusConflict, response)
		return
	}

	hashedPassword, err := utils.HashPassword(userReq.Password)
	if err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to hash password",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	user := models.User{
		Email:    userReq.Email,
		Password: hashedPassword,
		Name:     userReq.Name,
	}
	if err := uc.DB.Create(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to create user",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.CreateUserResponse{
		UserID: user.ID.String(),
		Email:  user.Email,
		Name:   user.Name,
	}

	c.JSON(http.StatusCreated, resultResponse)
}

// GetUsers godoc
// @Summary Get all users
// @Description Retrieve a list of all users
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.GetUsersResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /users [get]
func (uc *UserController) GetUsers(c *gin.Context) {
	var users []models.User
	if err := uc.DB.Find(&users).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to retrieve users",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.GetUsersResponse{
		Users: make([]dtos.GetUserResponse, len(users)),
	}
	for i, user := range users {
		resultResponse.Users[i] = dtos.GetUserResponse{
			UserID: user.ID.String(),
			Email:  user.Email,
			Name:   user.Name,
		}
	}

	c.JSON(http.StatusOK, resultResponse)
}

// GetUser godoc
// @Summary Get user by ID
// @Description Retrieve detailed information about the current user
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.GetUserResponseDetail
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Router /users/{id} [get]
func (uc *UserController) GetUser(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var user models.User
	if err := uc.DB.Model(&models.User{}).
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	resultResponse := dtos.GetUserResponseDetail{
		GetUserResponse: dtos.GetUserResponse{
			UserID: user.ID.String(),
			Email:  user.Email,
			Name:   user.Name,
		},
		Projects: user.Projects,
	}

	c.JSON(http.StatusOK, resultResponse)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update the current user's information
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user body dtos.UpdateUserRequest true "Updated user details"
// @Success 200 {object} dtos.UpdateUserResponse
// @Failure 400 {object} dtos.ErrorResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /users/{id} [patch]
func (uc *UserController) UpdateUser(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var user models.User
	if err := uc.DB.Model(&models.User{}).
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	var updateReq dtos.UpdateUserRequest
	if err := c.ShouldBindJSON(&updateReq); err != nil {
		response := dtos.ErrorResponse{
			Message: "Invalid request",
		}
		c.JSON(http.StatusBadRequest, response)
		return
	}

	if updateReq.Name != "" {
		user.Name = updateReq.Name
	}

	if updateReq.Password != "" {
		hashedPassword, err := utils.HashPassword(updateReq.Password)
		if err != nil {
			response := dtos.ErrorResponse{
				Message: "Failed to hash password",
			}
			c.JSON(http.StatusInternalServerError, response)
			return
		}
		user.Password = hashedPassword
	}

	if err := uc.DB.Save(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to update user",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.UpdateUserResponse{
		Email: user.Email,
		Name:  user.Name,
	}

	c.JSON(http.StatusOK, resultResponse)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete the current user account
// @Tags users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dtos.DeleteUserResponse
// @Failure 401 {object} dtos.ErrorResponse
// @Failure 404 {object} dtos.ErrorResponse
// @Failure 500 {object} dtos.ErrorResponse
// @Router /users/{id} [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	userID, exist := c.Get("user_id")
	if !exist {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusUnauthorized, response)
		return
	}

	var user models.User
	if err := uc.DB.Model(&models.User{}).
		Where("id = ?", userID).
		First(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "User not found",
		}
		c.JSON(http.StatusNotFound, response)
		return
	}

	if err := uc.DB.Delete(&user).Error; err != nil {
		response := dtos.ErrorResponse{
			Message: "Failed to delete user",
		}
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	resultResponse := dtos.DeleteUserResponse{
		Message: "User deleted successfully",
	}

	c.JSON(http.StatusOK, resultResponse)
}
