package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Wrk_Api/internal/handlers"
	"Wrk_Api/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCreateRetrospectiveItem(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "retro_owner@example.com", "Owner")
	projectId := createProjectForSprintTest(r, token)
	sprintId := createSprintForStoryTest(r, token, projectId)

	retroReq := handlers.CreateRetrospectiveItemRequest{
		Type:    "GOOD",
		Content: "Great teamwork!",
	}
	jsonValue, _ := json.Marshal(retroReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/sprints/"+sprintId+"/retrospectives/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var item models.RetrospectiveItem
	err := json.Unmarshal(w.Body.Bytes(), &item)
	assert.Nil(t, err)
	assert.Equal(t, "GOOD", item.Type)
	assert.Equal(t, "Great teamwork!", item.Content)
	assert.Equal(t, sprintId, item.SprintID)
}

func TestGetNotifications(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, userId := GetAuthToken(r, "notif_user@example.com", "Notif User")

	// Manually create notification using helper (simulating system event)
	err := handlers.CreateNotification(userId, "Test Alert", "Something happened", "INFO")
	assert.Nil(t, err)

	req, _ := http.NewRequest("GET", "/api/notifications/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var notifications []models.Notification
	json.Unmarshal(w.Body.Bytes(), &notifications)
	assert.Len(t, notifications, 1)
	assert.Equal(t, "Test Alert", notifications[0].Title)
	assert.False(t, notifications[0].Read)
}

func TestMarkNotificationRead(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, userId := GetAuthToken(r, "read_user@example.com", "Read User")

	handlers.CreateNotification(userId, "Unread", "Msg", "INFO")

	// Get ID
	req, _ := http.NewRequest("GET", "/api/notifications/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var notifications []models.Notification
	json.Unmarshal(w.Body.Bytes(), &notifications)
	notifId := notifications[0].ID

	// Mark Read
	req, _ = http.NewRequest("PUT", "/api/notifications/"+notifId+"/read", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify
	req, _ = http.NewRequest("GET", "/api/notifications/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	json.Unmarshal(w.Body.Bytes(), &notifications)
	assert.True(t, notifications[0].Read)
}
