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

func createSprintForStoryTest(r *gin.Engine, token string, projectId string) string {
	sprintReq := handlers.CreateSprintRequest{
		Name:      "Story Sprint",
		StartDate: time.Now(),
		EndDate:   time.Now().Add(time.Hour * 24 * 14),
	}
	jsonValue, _ := json.Marshal(sprintReq)
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/sprints/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var sprint models.Sprint
	json.Unmarshal(w.Body.Bytes(), &sprint)
	return sprint.ID
}

func TestCreateUserStory(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, userId := GetAuthToken(r, "story_owner@example.com", "Story Owner")
	projectId := createProjectForSprintTest(r, token) // Reusing helper
	sprintId := createSprintForStoryTest(r, token, projectId)

	storyReq := handlers.CreateUserStoryRequest{
		Title:       "My First Story",
		Description: "As a user...",
		Priority:    "HIGH",
		SprintID:    &sprintId,
		AssigneeID:  &userId,
	}
	jsonValue, _ := json.Marshal(storyReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/stories/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var story models.UserStory
	err := json.Unmarshal(w.Body.Bytes(), &story)
	assert.Nil(t, err)
	assert.Equal(t, "My First Story", story.Title)
	assert.Equal(t, "HIGH", story.Priority)
	assert.Equal(t, sprintId, *story.SprintID)
	assert.Equal(t, userId, *story.AssigneeID)
}

func TestUpdateUserStoryStatus(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "status_updater@example.com", "Updater")
	projectId := createProjectForSprintTest(r, token)

	// Create Story
	storyReq := handlers.CreateUserStoryRequest{
		Title:       "To Be Done",
		Description: "...",
	}
	jsonValue, _ := json.Marshal(storyReq)
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/stories/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var story models.UserStory
	json.Unmarshal(w.Body.Bytes(), &story)
	storyId := story.ID

	// Update Status to DONE
	newStatus := "DONE"
	updateReq := handlers.UpdateUserStoryRequest{
		Status: &newStatus,
	}
	jsonValue, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", "/api/projects/"+projectId+"/stories/"+storyId, bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedStory models.UserStory
	json.Unmarshal(w.Body.Bytes(), &updatedStory)
	assert.Equal(t, "DONE", updatedStory.Status)
	assert.NotNil(t, updatedStory.CompletedAt)
}

func TestUserStoryValidation(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "validator@example.com", "Validator")
	projectId := createProjectForSprintTest(r, token)

	badSprintId := "non-existent-sprint"
	storyReq := handlers.CreateUserStoryRequest{
		Title:       "Bad Sprint Story",
		Description: "...",
		SprintID:    &badSprintId,
	}
	jsonValue, _ := json.Marshal(storyReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/stories/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
