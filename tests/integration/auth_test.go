package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/atqamz/kogase-backend/models"
	"github.com/atqamz/kogase-backend/utils"
	"github.com/stretchr/testify/assert"
)

func TestRegisterUser(t *testing.T) {
	// Clean up before test
	testDB.Exec("DELETE FROM users")

	t.Run("RegisterValidUser", func(t *testing.T) {
		// Create a request payload
		reqBody := map[string]string{
			"name":     "Test User",
			"email":    "test@example.com",
			"password": "Password123",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to the register endpoint
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 200 (OK)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert that the response contains a token and user
		assert.Contains(t, response, "token")
		assert.Contains(t, response, "user")

		user, ok := response["user"].(map[string]interface{})
		assert.True(t, ok)

		// Assert that the user has the correct name and email
		assert.Equal(t, "Test User", user["name"])
		assert.Equal(t, "test@example.com", user["email"])
	})

	t.Run("RegisterDuplicateEmail", func(t *testing.T) {
		// Create a request payload
		reqBody := map[string]string{
			"name":     "Another User",
			"email":    "test@example.com", // Same email as before
			"password": "Password123",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to the register endpoint
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 400 (Bad Request) or 409 (Conflict)
		assert.True(t, rec.Code == http.StatusBadRequest || rec.Code == http.StatusConflict)

		// Parse the response
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert that the response contains an error
		assert.Contains(t, response, "error")
	})

	t.Run("RegisterInvalidEmail", func(t *testing.T) {
		// Create a request payload
		reqBody := map[string]string{
			"name":     "Invalid User",
			"email":    "notanemail", // Invalid email format
			"password": "Password123",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to the register endpoint
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 400 (Bad Request)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		// Parse the response
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert that the response contains an error
		assert.Contains(t, response, "error")
	})
}

func TestLoginUser(t *testing.T) {
	// Clean up before test
	testDB.Exec("DELETE FROM users")

	// Create a test user
	// First, hash the password
	password := "Password123"
	hashedPassword, err := utils.HashPassword(password)
	assert.NoError(t, err)

	// Then, create the user
	user := models.User{
		Name:     "Login Test User",
		Email:    "login@example.com",
		Password: hashedPassword,
	}
	result := testDB.Create(&user)
	assert.NoError(t, result.Error)

	t.Run("LoginValidUser", func(t *testing.T) {
		// Create a request payload
		reqBody := map[string]string{
			"email":    "login@example.com",
			"password": "Password123",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to the login endpoint
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 200 (OK)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert that the response contains a token and user
		assert.Contains(t, response, "token")
		assert.Contains(t, response, "user")

		responseUser, ok := response["user"].(map[string]interface{})
		assert.True(t, ok)

		// Assert that the user has the correct name and email
		assert.Equal(t, "Login Test User", responseUser["name"])
		assert.Equal(t, "login@example.com", responseUser["email"])
	})

	t.Run("LoginInvalidPassword", func(t *testing.T) {
		// Create a request payload
		reqBody := map[string]string{
			"email":    "login@example.com",
			"password": "WrongPassword",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to the login endpoint
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 401 (Unauthorized)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		// Parse the response
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert that the response contains an error
		assert.Contains(t, response, "error")
	})

	t.Run("LoginNonExistentUser", func(t *testing.T) {
		// Create a request payload
		reqBody := map[string]string{
			"email":    "nonexistent@example.com",
			"password": "Password123",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to the login endpoint
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 401 (Unauthorized)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)

		// Parse the response
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert that the response contains an error
		assert.Contains(t, response, "error")
	})
}
