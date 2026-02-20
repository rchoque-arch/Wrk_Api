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

func TestCreateTask(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, userId := GetAuthToken(r, "task_owner@example.com", "Task Owner")
	projectId := createProjectForSprintTest(r, token)
	sprintId := createSprintForStoryTest(r, token, projectId)

	taskReq := handlers.CreateTaskRequest{
		Title:       "My First Task",
		Description: nil,
		Priority:    "HIGH",
		SprintID:    &sprintId,
		AssigneeID:  &userId,
	}
	jsonValue, _ := json.Marshal(taskReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/tasks/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var task models.Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	assert.Nil(t, err)
	assert.Equal(t, "My First Task", task.Title)
	assert.Equal(t, "HIGH", task.Priority)
	assert.Equal(t, sprintId, *task.SprintID)
	assert.Equal(t, userId, *task.AssigneeID)
}

func TestUpdateTaskStatus(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "task_updater@example.com", "Updater")
	projectId := createProjectForSprintTest(r, token)

	// Create Task
	taskReq := handlers.CreateTaskRequest{
		Title: "To Do Task",
	}
	jsonValue, _ := json.Marshal(taskReq)
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/tasks/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var task models.Task
	json.Unmarshal(w.Body.Bytes(), &task)
	taskId := task.ID

	// Update Status to DONE
	newStatus := "DONE"
	updateReq := handlers.UpdateTaskRequest{
		Status: &newStatus,
	}
	jsonValue, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", "/api/projects/"+projectId+"/tasks/"+taskId, bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedTask models.Task
	json.Unmarshal(w.Body.Bytes(), &updatedTask)
	assert.Equal(t, "DONE", updatedTask.Status)
	assert.NotNil(t, updatedTask.CompletedAt)
}

func TestTaskValidation(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "task_validator@example.com", "Validator")
	projectId := createProjectForSprintTest(r, token)

	badSprintId := "non-existent-sprint"
	taskReq := handlers.CreateTaskRequest{
		Title:    "Bad Sprint Task",
		SprintID: &badSprintId,
	}
	jsonValue, _ := json.Marshal(taskReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/tasks/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
