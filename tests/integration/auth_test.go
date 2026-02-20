package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Wrk_Api/internal/handlers"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	user := handlers.RegisterRequest{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "password123",
	}
	jsonValue, _ := json.Marshal(user)

	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response handlers.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, user.Email, response.User.Email)
}

func TestLogin(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	// Create user first
	password := "password123"
	registerReq := handlers.RegisterRequest{
		Email:    "login@example.com",
		Name:     "Login User",
		Password: password,
	}
	jsonValue, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Test Login
	loginReq := handlers.LoginRequest{
		Email:    "login@example.com",
		Password: password,
	}
	jsonValue, _ = json.Marshal(loginReq)
	req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonValue))
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response handlers.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.Nil(t, err)
	assert.NotEmpty(t, response.Token)
	assert.Equal(t, registerReq.Email, response.User.Email)
}
