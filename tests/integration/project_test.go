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

func TestCreateProject(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "owner@example.com", "Owner")

	projectReq := handlers.CreateProjectRequest{
		Name: "New Project",
	}
	jsonValue, _ := json.Marshal(projectReq)

	req, _ := http.NewRequest("POST", "/api/projects/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var project models.Project
	err := json.Unmarshal(w.Body.Bytes(), &project)
	assert.Nil(t, err)
	assert.Equal(t, projectReq.Name, project.Name)
	assert.NotEmpty(t, project.ID)
	assert.NotEmpty(t, project.OwnerID)
}

func TestGetProjects(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "user@example.com", "User")

	// Create a project first
	projectReq := handlers.CreateProjectRequest{
		Name: "My Project",
	}
	jsonValue, _ := json.Marshal(projectReq)
	req, _ := http.NewRequest("POST", "/api/projects/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Get projects
	req, _ = http.NewRequest("GET", "/api/projects/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var projects []models.Project
	err := json.Unmarshal(w.Body.Bytes(), &projects)
	assert.Nil(t, err)
	assert.Len(t, projects, 1)
	assert.Equal(t, "My Project", projects[0].Name)
}

func TestProjectAccessControl(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	ownerToken, _ := GetAuthToken(r, "owner_unique@example.com", "Owner")
	otherToken, _ := GetAuthToken(r, "other_unique@example.com", "Other")

	// Owner creates project
	projectReq := handlers.CreateProjectRequest{Name: "Private Project"}
	jsonValue, _ := json.Marshal(projectReq)
	req, _ := http.NewRequest("POST", "/api/projects/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var project models.Project
	json.Unmarshal(w.Body.Bytes(), &project)
	projectId := project.ID

	// Other user tries to update project
	updateReq := handlers.UpdateProjectRequest{Name: "Hacked Project"}
	jsonValue, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", "/api/projects/"+projectId, bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+otherToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	// Owner updates project
	jsonValue, _ = json.Marshal(updateReq) // Marshal again to be safe
	req, _ = http.NewRequest("PUT", "/api/projects/"+projectId, bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+ownerToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
