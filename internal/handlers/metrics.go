package handlers

import (
	"net/http"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"github.com/gin-gonic/gin"
)

const statusDone = "DONE"

type ProjectMetrics struct {
	TotalTasks       int64            `json:"totalTasks"`
	CompletedTasks   int64            `json:"completedTasks"`
	TaskStatusCounts map[string]int64 `json:"taskStatusCounts"`
	TotalPoints      int              `json:"totalPoints"`
	CompletedPoints  int              `json:"completedPoints"`
	SprintVelocity   float64          `json:"sprintVelocity"` // Average points per completed sprint
}

func GetProjectMetrics(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)
	projectId := c.Param("projectId")

	if !isProjectMember(userId, projectId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	metrics := ProjectMetrics{
		TaskStatusCounts: make(map[string]int64),
	}

	// 1. Task Statistics
	var tasks []models.Task
	if err := database.DB.Where("project_id = ?", projectId).Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	metrics.TotalTasks = int64(len(tasks))
	for _, t := range tasks {
		metrics.TaskStatusCounts[t.Status]++
		if t.Status == statusDone {
			metrics.CompletedTasks++
		}
	}

	// 2. User Story Points (Velocity)
	var stories []models.UserStory
	if err := database.DB.Where("project_id = ?", projectId).Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user stories"})
		return
	}

	for _, s := range stories {
		points := 0
		if s.StoryPoints != nil {
			points = *s.StoryPoints
		}
		metrics.TotalPoints += points
		if s.Status == statusDone {
			metrics.CompletedPoints += points
		}
	}

	// 3. Sprint Velocity (Average points of completed sprints)
	// Find sprints that are essentially "done" (e.g. end date passed or status completed)
	// For simplicity, let's assume we calculate based on stories linked to sprints.
	// Query: Select sprint_id, sum(story_points) group by sprint_id where status='DONE'
	// Simplified logic: iterate stories
	sprintPoints := make(map[string]int)
	completedSprints := make(map[string]bool)

	for _, s := range stories {
		if s.SprintID != nil && s.Status == statusDone {
			points := 0
			if s.StoryPoints != nil {
				points = *s.StoryPoints
			}
			sprintPoints[*s.SprintID] += points
			completedSprints[*s.SprintID] = true
		}
	}

	if len(completedSprints) > 0 {
		totalVelocity := 0
		for _, p := range sprintPoints {
			totalVelocity += p
		}
		metrics.SprintVelocity = float64(totalVelocity) / float64(len(completedSprints))
	}

	c.JSON(http.StatusOK, metrics)
}
