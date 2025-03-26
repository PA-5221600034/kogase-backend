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

func setupTestUser(t *testing.T) (models.User, string) {
	// Clean up before test
	testDB.Exec("DELETE FROM auth_tokens")
	testDB.Exec("DELETE FROM projects")
	testDB.Exec("DELETE FROM users")

	// Create a test user
	password := "Password123"
	hashedPassword, err := utils.HashPassword(password)
	assert.NoError(t, err)

	user := models.User{
		Name:     "Project Test User",
		Email:    "project@example.com",
		Password: hashedPassword,
	}
	result := testDB.Create(&user)
	assert.NoError(t, result.Error)

	// Log in to get an auth token
	reqBody := map[string]string{
		"email":    "project@example.com",
		"password": "Password123",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response, "token")

	token := response["token"].(string)
	return user, token
}

func TestProjectCRUD(t *testing.T) {
	_, token := setupTestUser(t)

	var projectID string

	t.Run("CreateProject", func(t *testing.T) {
		// Create a request payload
		reqBody := map[string]string{
			"name":        "Test Project",
			"description": "This is a test project",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to create a project
		req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboard/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 201 (Created) or 200 (OK)
		assert.True(t, rec.Code == http.StatusCreated || rec.Code == http.StatusOK)

		// Parse the response
		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Store the project ID for later tests
		projectID = response["id"].(string)

		// Assert that the project has the correct name and description
		assert.Equal(t, "Test Project", response["name"])
		assert.Equal(t, "This is a test project", response["description"])
	})

	t.Run("GetProjects", func(t *testing.T) {
		// Create a request to get all projects
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/projects", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 200 (OK)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response
		var response []map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert that the response contains the project we created
		assert.GreaterOrEqual(t, len(response), 1)
		found := false
		for _, project := range response {
			if project["id"] == projectID {
				found = true
				assert.Equal(t, "Test Project", project["name"])
				assert.Equal(t, "This is a test project", project["description"])
				break
			}
		}
		assert.True(t, found, "Project not found in the list")
	})

	t.Run("GetProject", func(t *testing.T) {
		// Create a request to get a specific project
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/projects/"+projectID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 200 (OK)
		assert.Equal(t, http.StatusOK, rec.Code)

		// Parse the response
		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Assert that the project has the correct name and description
		assert.Equal(t, projectID, response["id"])
		assert.Equal(t, "Test Project", response["name"])
		assert.Equal(t, "This is a test project", response["description"])
	})

	t.Run("UpdateProject", func(t *testing.T) {
		// Create a request payload
		reqBody := map[string]string{
			"name":        "Updated Project",
			"description": "This project has been updated",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to update a project
		req := httptest.NewRequest(http.MethodPut, "/api/v1/dashboard/projects/"+projectID, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

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

		// Assert that the project has been updated
		assert.Equal(t, projectID, response["id"])
		assert.Equal(t, "Updated Project", response["name"])
		assert.Equal(t, "This project has been updated", response["description"])
	})

	t.Run("RegenerateAPIKey", func(t *testing.T) {
		// First, get the current API key
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/projects/"+projectID+"/api-key", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		var response map[string]interface{}
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		originalAPIKey := response["api_key"].(string)
		assert.NotEmpty(t, originalAPIKey)

		// Now regenerate the API key
		req = httptest.NewRequest(http.MethodPost, "/api/v1/dashboard/projects/"+projectID+"/api-key/regenerate", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)

		newAPIKey := response["api_key"].(string)
		assert.NotEmpty(t, newAPIKey)
		assert.NotEqual(t, originalAPIKey, newAPIKey, "API key should have changed")
	})

	t.Run("DeleteProject", func(t *testing.T) {
		// Create a request to delete a project
		req := httptest.NewRequest(http.MethodDelete, "/api/v1/dashboard/projects/"+projectID, nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 200 (OK) or 204 (No Content)
		assert.True(t, rec.Code == http.StatusOK || rec.Code == http.StatusNoContent)

		// Now try to get the project and it should return 404
		req = httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/projects/"+projectID, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec = httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNotFound, rec.Code)
	})
}

func TestUnauthorizedAccess(t *testing.T) {
	t.Run("NoToken", func(t *testing.T) {
		// Create a request without a token
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/projects", nil)
		rec := httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("InvalidToken", func(t *testing.T) {
		// Create a request with an invalid token
		req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/projects", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rec := httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("AccessOtherUserProject", func(t *testing.T) {
		// Create two users
		_, token1 := setupTestUser(t)

		// Create a project for user 1
		reqBody := map[string]string{
			"name":        "User 1 Project",
			"description": "This project belongs to user 1",
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboard/projects", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token1)
		rec := httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)
		assert.True(t, rec.Code == http.StatusCreated || rec.Code == http.StatusOK)

		var response map[string]interface{}
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		projectID := response["id"].(string)

		// Now create a second user
		testDB.Exec("DELETE FROM auth_tokens WHERE 1=1")
		testDB.Exec("DELETE FROM users WHERE email = ?", "project2@example.com")

		password := "Password123"
		hashedPassword, err := utils.HashPassword(password)
		assert.NoError(t, err)

		user2 := models.User{
			Name:     "Project Test User 2",
			Email:    "project2@example.com",
			Password: hashedPassword,
		}
		result := testDB.Create(&user2)
		assert.NoError(t, result.Error)

		reqBody = map[string]string{
			"email":    "project2@example.com",
			"password": "Password123",
		}
		body, err = json.Marshal(reqBody)
		assert.NoError(t, err)

		req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		err = json.Unmarshal(rec.Body.Bytes(), &response)
		assert.NoError(t, err)
		token2 := response["token"].(string)

		// Try to access user 1's project with user 2's token
		req = httptest.NewRequest(http.MethodGet, "/api/v1/dashboard/projects/"+projectID, nil)
		req.Header.Set("Authorization", "Bearer "+token2)
		rec = httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)

		// Should return 403 Forbidden or 404 Not Found
		assert.True(t, rec.Code == http.StatusForbidden || rec.Code == http.StatusNotFound)
	})
}
