package handlers

import (
	"net/http"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EvaluationCriteriaInput struct {
	CriteriaID string  `json:"criteriaId" binding:"required"`
	Score      int     `json:"score" binding:"required"`
	Comment    *string `json:"comment"`
}

type CreateEvaluationRequest struct {
	TaskID   *string                   `json:"taskId"`
	SprintID *string                   `json:"sprintId"`
	Criteria []EvaluationCriteriaInput `json:"criteria" binding:"required,dive"`
	Feedback *string                   `json:"feedback"`
}

func CreateEvaluation(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var req CreateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.TaskID == nil && req.SprintID == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Evaluation must be linked to a task or sprint"})
		return
	}

	// Validate Task/Sprint belongs to project
	if req.TaskID != nil {
		var count int64
		database.DB.Model(&models.Task{}).Where("id = ? AND project_id = ?", *req.TaskID, projectID).Count(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
			return
		}
	} else if req.SprintID != nil {
		var count int64
		database.DB.Model(&models.Sprint{}).Where("id = ? AND project_id = ?", *req.SprintID, projectID).Count(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
			return
		}
	}

	evaluationId := uuid.NewString()
	evaluation := models.Evaluation{
		ID:          evaluationId,
		ProjectID:   projectID,
		TaskID:      req.TaskID,
		SprintID:    req.SprintID,
		EvaluatorID: userID,
		Status:      "COMPLETED",
		Feedback:    req.Feedback,
	}

	// Calculate Score and Build Criteria
	totalScore := 0
	totalMaxScore := 0
	var evalCriteriaList []models.EvaluationCriteria

	for _, item := range req.Criteria {
		// Verify criteria exists (optimally should fetch all valid criteria IDs first)
		var criteria models.Criteria
		if err := database.DB.First(&criteria, "id = ?", item.CriteriaID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid criteria ID: " + item.CriteriaID})
			return
		}

		if item.Score > criteria.MaxScore {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Score exceeds max score for criteria: " + criteria.Name})
			return
		}

		totalScore += item.Score
		totalMaxScore += criteria.MaxScore

		evalCriteriaList = append(evalCriteriaList, models.EvaluationCriteria{
			ID:           uuid.NewString(),
			EvaluationID: evaluationId,
			CriteriaID:   item.CriteriaID,
			Score:        item.Score,
			Comment:      item.Comment,
		})
	}

	// Normalize score to 100? Or keep raw sum? Let's keep raw sum for now or normalized average.
	// Simple sum for now.
	score := totalScore
	evaluation.Score = &score
	evaluation.Criteria = evalCriteriaList

	if err := database.DB.Create(&evaluation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit evaluation"})
		return
	}

	c.JSON(http.StatusCreated, evaluation)
}

func GetEvaluations(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var evaluations []models.Evaluation
	query := database.DB.Where("project_id = ?", projectID)

	if taskID := c.Query("taskId"); taskID != "" {
		query = query.Where("task_id = ?", taskID)
	}
	if sprintID := c.Query("sprintId"); sprintID != "" {
		query = query.Where("sprint_id = ?", sprintID)
	}

	if err := query.Preload("Evaluator").Preload("Criteria").Preload("Criteria.Criteria").Find(&evaluations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch evaluations"})
		return
	}

	c.JSON(http.StatusOK, evaluations)
}
