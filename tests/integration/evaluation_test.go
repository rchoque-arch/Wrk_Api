package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"Wrk_Api/internal/handlers"
	"Wrk_Api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func createRubricForTest(r *gin.Engine, token, projectId string) string {
	rubricReq := handlers.CreateRubricRequest{
		Name: "Code Review",
		Criteria: []handlers.CreateCriteriaRequest{
			{Name: "Quality", MaxScore: 10, Weight: 1},
			{Name: "Performance", MaxScore: 10, Weight: 1},
		},
	}
	jsonValue, _ := json.Marshal(rubricReq)
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/rubrics/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var rubric models.Rubric
	json.Unmarshal(w.Body.Bytes(), &rubric)
	return rubric.ID
}

func getCriteriaIds(r *gin.Engine, token, projectId, rubricId string) []string {
	req, _ := http.NewRequest("GET", "/api/projects/"+projectId+"/rubrics/"+rubricId, nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var rubric models.Rubric
	json.Unmarshal(w.Body.Bytes(), &rubric)

	var ids []string
	for _, c := range rubric.Criteria {
		ids = append(ids, c.ID)
	}
	return ids
}

func createTaskForTest(r *gin.Engine, token, projectId string) string {
	taskReq := handlers.CreateTaskRequest{Title: "Task for Eval"}
	jsonValue, _ := json.Marshal(taskReq)
	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/tasks/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var task models.Task
	json.Unmarshal(w.Body.Bytes(), &task)
	return task.ID
}

func TestCreateRubric(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "eval_owner@example.com", "Owner")
	projectId := createProjectForSprintTest(r, token)

	rubricReq := handlers.CreateRubricRequest{
		Name: "Peer Review",
		Criteria: []handlers.CreateCriteriaRequest{
			{Name: "Communication", MaxScore: 5},
		},
	}
	jsonValue, _ := json.Marshal(rubricReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/rubrics/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var rubric models.Rubric
	err := json.Unmarshal(w.Body.Bytes(), &rubric)
	assert.Nil(t, err)
	assert.Equal(t, "Peer Review", rubric.Name)
	assert.Len(t, rubric.Criteria, 1)
	assert.Equal(t, "Communication", rubric.Criteria[0].Name)
}

func TestCreateEvaluation(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "evaluator@example.com", "Evaluator")
	projectId := createProjectForSprintTest(r, token)
	taskId := createTaskForTest(r, token, projectId)
	rubricId := createRubricForTest(r, token, projectId)
	criteriaIds := getCriteriaIds(r, token, projectId, rubricId)

	assert.Len(t, criteriaIds, 2)

	evalReq := handlers.CreateEvaluationRequest{
		TaskID: &taskId,
		Criteria: []handlers.EvaluationCriteriaInput{
			{CriteriaID: criteriaIds[0], Score: 8}, // Quality
			{CriteriaID: criteriaIds[1], Score: 9}, // Performance
		},
	}
	jsonValue, _ := json.Marshal(evalReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/evaluations/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var eval models.Evaluation
	err := json.Unmarshal(w.Body.Bytes(), &eval)
	assert.Nil(t, err)
	assert.Equal(t, 17, *eval.Score)
	assert.Len(t, eval.Criteria, 2)
}

func TestEvaluationScoreValidation(t *testing.T) {
	SetupTestDB()
	r := SetupRouter()

	token, _ := GetAuthToken(r, "bad_evaluator@example.com", "Evaluator")
	projectId := createProjectForSprintTest(r, token)
	taskId := createTaskForTest(r, token, projectId)
	rubricId := createRubricForTest(r, token, projectId)
	criteriaIds := getCriteriaIds(r, token, projectId, rubricId)

	evalReq := handlers.CreateEvaluationRequest{
		TaskID: &taskId,
		Criteria: []handlers.EvaluationCriteriaInput{
			{CriteriaID: criteriaIds[0], Score: 11}, // Max is 10
		},
	}
	jsonValue, _ := json.Marshal(evalReq)

	req, _ := http.NewRequest("POST", "/api/projects/"+projectId+"/evaluations/", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
