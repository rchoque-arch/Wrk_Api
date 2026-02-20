package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"Wrk_Api/internal/handlers"
	"Wrk_Api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func createProjectForSprintTest(r *gin.Engine, token string) string {
	projectReq := handlers.CreateProjectRequest{
		Name: "Sprint Project",
	}
	jsonValue, _ := json.Marshal(projectReq)
	req, _ := http.NewRequest("POST", "/api/projects/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var project models.Project
	json.Unmarshal(w.Body.Bytes(), &project)
	return project.ID
}

func TestCreateSprint(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "sprint_owner@example.com", "Sprint Owner")
	projectId := createProjectForSprintTest(r, token)

	startDate := time.Now()
	endDate := startDate.Add(time.Hour * 24 * 14) // 2 weeks

	sprintReq := handlers.CreateSprintRequest{
		Name:      "Sprint 1",
		StartDate: startDate,
		EndDate:   endDate,
	}
	jsonValue, _ := json.Marshal(sprintReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/sprints/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var sprint models.Sprint
	err := json.Unmarshal(w.Body.Bytes(), &sprint)
	assert.Nil(t, err)
	assert.Equal(t, "Sprint 1", sprint.Name)
	assert.Equal(t, projectId, sprint.ProjectID)
}

func TestCreateSprintInvalidDates(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "sprint_dates@example.com", "Sprint User")
	projectId := createProjectForSprintTest(r, token)

	startDate := time.Now()
	endDate := startDate.Add(-time.Hour * 24) // Yesterday

	sprintReq := handlers.CreateSprintRequest{
		Name:      "Bad Sprint",
		StartDate: startDate,
		EndDate:   endDate,
	}
	jsonValue, _ := json.Marshal(sprintReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/sprints/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestSprintAccessControl(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	ownerToken, _ := GetAuthToken(r, "sprint_ac_owner@example.com", "Owner")
	otherToken, _ := GetAuthToken(r, "sprint_ac_other@example.com", "Other")

	projectId := createProjectForSprintTest(r, ownerToken)

	// Other user tries to create sprint
	sprintReq := handlers.CreateSprintRequest{
		Name:      "Hacked Sprint",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(time.Hour * 24),
	}
	jsonValue, _ := json.Marshal(sprintReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/sprints/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+otherToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
