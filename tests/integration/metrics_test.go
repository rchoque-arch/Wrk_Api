package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Wrk_Api/internal/handlers"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetProjectMetrics(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "metrics_owner@example.com", "Owner")
	projectId := createProjectForSprintTest(r, token)
	sprintId := createSprintForStoryTest(r, token, projectId)

	// Create some Tasks
	createTask(r, token, projectId, "Task 1", "TODO")
	createTask(r, token, projectId, "Task 2", "DONE")
	createTask(r, token, projectId, "Task 3", "IN_PROGRESS")

	// Create some User Stories with Points
	createStory(r, token, projectId, sprintId, 5, "DONE")
	createStory(r, token, projectId, sprintId, 3, "TODO")

	// Fetch Metrics
	req, _ := http.NewRequest("GET", "/api/projects/"+projectId+"/metrics", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var metrics handlers.ProjectMetrics
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	assert.Nil(t, err)

	// Verify Tasks
	assert.Equal(t, int64(3), metrics.TotalTasks)
	assert.Equal(t, int64(1), metrics.CompletedTasks) // 1 DONE
	assert.Equal(t, int64(1), metrics.TaskStatusCounts["TODO"])
	assert.Equal(t, int64(1), metrics.TaskStatusCounts["DONE"])
	assert.Equal(t, int64(1), metrics.TaskStatusCounts["IN_PROGRESS"])

	// Verify Points
	assert.Equal(t, 8, metrics.TotalPoints)     // 5 + 3
	assert.Equal(t, 5, metrics.CompletedPoints) // 5 (DONE)

	// Verify Velocity (1 completed sprint with 5 points)
	assert.Equal(t, 5.0, metrics.SprintVelocity)
}

func createTask(r *gin.Engine, token, projectId, title, status string) {
	req := handlers.CreateTaskRequest{
		Title: title,
	}
	jsonValue, _ := json.Marshal(req)

	// Create
	httpReq, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/tasks/", bytes.NewBuffer(jsonValue))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)

	// Update status if needed
	if status != "TODO" {
		var task map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &task)
		taskId := task["id"].(string)

		updateReq := handlers.UpdateTaskRequest{Status: &status}
		jsonValue, _ = json.Marshal(updateReq)
		httpReq, _ = http.NewRequest("PUT", "/api/projects/"+projectId+"/tasks/"+taskId, bytes.NewBuffer(jsonValue))
		httpReq.Header.Set("Authorization", "Bearer "+token)
		w = httptest.NewRecorder()
		r.ServeHTTP(w, httpReq)
	}
}

func createStory(r *gin.Engine, token, projectId, sprintId string, points int, status string) {
	req := handlers.CreateUserStoryRequest{
		Title:       "Story",
		Description: "Desc",
		StoryPoints: &points,
		SprintID:    &sprintId,
	}
	jsonValue, _ := json.Marshal(req)

	// Create
	httpReq, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/stories/", bytes.NewBuffer(jsonValue))
	httpReq.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)

	// Update status if needed
	if status != "BACKLOG" { // Default is BACKLOG
		var story map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &story)
		if idVal, ok := story["id"].(string); ok {
			storyId := idVal

			updateReq := handlers.UpdateUserStoryRequest{Status: &status}
			jsonValue, _ = json.Marshal(updateReq)
			httpReq, _ = http.NewRequest("PUT", "/api/projects/"+projectId+"/stories/"+storyId, bytes.NewBuffer(jsonValue))
			httpReq.Header.Set("Authorization", "Bearer "+token)
			w = httptest.NewRecorder()
			r.ServeHTTP(w, httpReq)
		}
	}
}
