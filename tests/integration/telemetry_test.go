package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/atqamz/kogase-backend/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupTestProjectWithAPIKey(t *testing.T) (models.Project, string) {
	// Clean up before test
	testDB.Exec("DELETE FROM auth_tokens")
	testDB.Exec("DELETE FROM events")
	testDB.Exec("DELETE FROM devices")
	testDB.Exec("DELETE FROM project_users")
	testDB.Exec("DELETE FROM projects")
	testDB.Exec("DELETE FROM users")

	// Setup user and get token
	_, token := setupTestUser(t)

	// Create a project for the user
	reqBody := map[string]string{
		"name":        "Telemetry Test Project",
		"description": "Project for testing telemetry",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/dashboard/projects", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.True(t, rec.Code == http.StatusCreated || rec.Code == http.StatusOK)

	var projectResponse map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &projectResponse)
	assert.NoError(t, err)
	projectID := projectResponse["id"].(string)

	// Get the API key
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/dashboard/projects/%s/api-key", projectID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	var apiKeyResponse map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &apiKeyResponse)
	assert.NoError(t, err)
	apiKey := apiKeyResponse["api_key"].(string)

	// Get the project from DB
	var project models.Project
	result := testDB.First(&project, "id = ?", projectID)
	assert.NoError(t, result.Error)

	return project, apiKey
}

func TestTelemetryEvents(t *testing.T) {
	project, apiKey := setupTestProjectWithAPIKey(t)

	t.Run("RecordSingleEvent", func(t *testing.T) {
		// Create a request payload
		deviceID := uuid.New().String()
		reqBody := map[string]interface{}{
			"client_time": time.Now().Format(time.RFC3339),
			"event_name":  "test_event",
			"event_type":  "custom",
			"device_id":   deviceID,
			"platform":    "Unity",
			"os_version":  "Windows 10",
			"app_version": "1.0.0",
			"parameters": map[string]interface{}{
				"param1": "value1",
				"param2": 123,
			},
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to record an event
		req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/event", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 201 (Created) or 200 (OK)
		assert.True(t, rec.Code == http.StatusCreated || rec.Code == http.StatusOK)

		// Verify that the event was recorded in the database
		var count int64
		result := testDB.Model(&models.Event{}).Where("project_id = ? AND event_name = ?", project.ID, "test_event").Count(&count)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("RecordMultipleEvents", func(t *testing.T) {
		// Create a request payload with multiple events
		deviceID := uuid.New().String()
		reqBody := map[string]interface{}{
			"events": []map[string]interface{}{
				{
					"client_time": time.Now().Add(-10 * time.Minute).Format(time.RFC3339),
					"event_name":  "event_1",
					"event_type":  "custom",
					"device_id":   deviceID,
					"platform":    "Unity",
					"os_version":  "Windows 10",
					"app_version": "1.0.0",
					"parameters": map[string]interface{}{
						"param1": "batch1",
					},
				},
				{
					"client_time": time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
					"event_name":  "event_2",
					"event_type":  "custom",
					"device_id":   deviceID,
					"platform":    "Unity",
					"os_version":  "Windows 10",
					"app_version": "1.0.0",
					"parameters": map[string]interface{}{
						"param2": "batch2",
					},
				},
			},
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to record multiple events
		req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/events", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 201 (Created) or 200 (OK)
		assert.True(t, rec.Code == http.StatusCreated || rec.Code == http.StatusOK)

		// Verify that both events were recorded in the database
		var count int64
		var deviceUUID uuid.UUID

		// First find the device in the database
		var device models.Device
		result := testDB.Where("project_id = ? AND device_id = ?", project.ID, deviceID).First(&device)
		assert.NoError(t, result.Error)
		deviceUUID = device.ID

		// Then count events for this device
		result = testDB.Model(&models.Event{}).Where("project_id = ? AND device_id = ?", project.ID, deviceUUID).Count(&count)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(2), count)
	})

	t.Run("RecordSession", func(t *testing.T) {
		// Create a request payload for session start
		deviceID := uuid.New().String()
		sessionID := uuid.New().String()
		reqBody := map[string]interface{}{
			"client_time": time.Now().Format(time.RFC3339),
			"event_name":  sessionID, // Using session ID as event name
			"event_type":  "session_start",
			"device_id":   deviceID,
			"platform":    "Unity",
			"os_version":  "Windows 10",
			"app_version": "1.0.0",
			"parameters": map[string]interface{}{
				"session_id": sessionID,
			},
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to start a session
		req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/session/start", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 201 (Created) or 200 (OK)
		assert.True(t, rec.Code == http.StatusCreated || rec.Code == http.StatusOK)

		// Verify that the session start event was recorded in the database
		var count int64
		result := testDB.Model(&models.Event{}).
			Where("project_id = ? AND event_type = ?", project.ID, "session_start").
			Count(&count)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), count)

		// Now end the session
		reqBody = map[string]interface{}{
			"client_time": time.Now().Format(time.RFC3339),
			"event_name":  sessionID, // Using session ID as event name
			"event_type":  "session_end",
			"device_id":   deviceID,
			"platform":    "Unity",
			"os_version":  "Windows 10",
			"app_version": "1.0.0",
			"parameters": map[string]interface{}{
				"session_id": sessionID,
				"duration":   300, // 5 minutes in seconds
			},
		}
		body, err = json.Marshal(reqBody)
		assert.NoError(t, err)

		req = httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/session/end", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)
		rec = httptest.NewRecorder()
		testRouter.ServeHTTP(rec, req)
		assert.True(t, rec.Code == http.StatusCreated || rec.Code == http.StatusOK)

		// Verify that the session end event was recorded in the database
		result = testDB.Model(&models.Event{}).
			Where("project_id = ? AND event_type = ?", project.ID, "session_end").
			Count(&count)
		assert.NoError(t, result.Error)
		assert.Equal(t, int64(1), count)
	})

	t.Run("InvalidAPIKey", func(t *testing.T) {
		// Create a request payload
		reqBody := map[string]interface{}{
			"client_time": time.Now().Format(time.RFC3339),
			"event_name":  "test_event",
			"event_type":  "custom",
			"device_id":   uuid.New().String(),
			"platform":    "Unity",
			"os_version":  "Windows 10",
			"app_version": "1.0.0",
			"parameters":  map[string]interface{}{},
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request with an invalid API key
		req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/event", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "invalid-api-key")

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 401 (Unauthorized)
		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("MissingRequiredFields", func(t *testing.T) {
		// Create a request payload missing required fields
		reqBody := map[string]interface{}{
			// Missing client_time, event_name, and device_id
			"parameters": map[string]interface{}{
				"param1": "value1",
			},
		}
		body, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		// Create a request to record an event
		req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/event", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", apiKey)

		// Create a response recorder
		rec := httptest.NewRecorder()

		// Serve the request
		testRouter.ServeHTTP(rec, req)

		// Assert that the response status code is 400 (Bad Request)
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
