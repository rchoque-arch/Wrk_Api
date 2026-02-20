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

func TestGetUserProfile(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "profile@example.com", "My Profile")

	req, _ := http.NewRequest("GET", "/api/users/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user models.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.Nil(t, err)
	assert.Equal(t, "profile@example.com", user.Email)
	assert.Equal(t, "My Profile", user.Name)
}

func TestUpdateUserProfile(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "update_profile@example.com", "Old Name")

	newName := "New Name"
	reqBody := handlers.UpdateUserRequest{
		Name: &newName,
	}
	jsonValue, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("PUT", "/api/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var user models.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.Nil(t, err)
	assert.Equal(t, "New Name", user.Name)
}
