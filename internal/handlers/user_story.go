package handlers

import (
	"net/http"
	"time"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateUserStoryRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Acceptance  *string `json:"acceptance"`
	Priority    string  `json:"priority"` // MEDIUM
	StoryPoints *int    `json:"storyPoints"`
	SprintID    *string `json:"sprintId"`
	AssigneeID  *string `json:"assigneeId"`
}

type UpdateUserStoryRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Acceptance  *string `json:"acceptance"`
	Priority    *string `json:"priority"`
	StoryPoints *int    `json:"storyPoints"`
	Status      *string `json:"status"`
	SprintID    *string `json:"sprintId"`
	AssigneeID  *string `json:"assigneeId"`
}

func CreateUserStory(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")

	// Validate Access
	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var req CreateUserStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate Sprint if provided
	if req.SprintID != nil {
		var sprint models.Sprint
		if err := database.DB.First(&sprint, "id = ? AND project_id = ?", *req.SprintID, projectID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
			return
		}
	}

	// Validate Assignee if provided
	if req.AssigneeID != nil {
		if !isProjectMember(*req.AssigneeID, projectID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Assignee must be a project member"})
			return
		}
	}

	story := models.UserStory{
		ID:          uuid.NewString(),
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		Acceptance:  req.Acceptance,
		Priority:    "MEDIUM",
		StoryPoints: req.StoryPoints,
		Status:      "BACKLOG",
		SprintID:    req.SprintID,
		AssigneeID:  req.AssigneeID,
	}

	if req.Priority != "" {
		story.Priority = req.Priority
	}

	if err := database.DB.Create(&story).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user story"})
		return
	}

	c.JSON(http.StatusCreated, story)
}

func GetUserStories(c *gin.Context) {
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

	var stories []models.UserStory
	query := database.DB.Where("project_id = ?", projectID)

	// Optional filtering by Sprint
	if sprintID := c.Query("sprintId"); sprintID != "" {
		query = query.Where("sprint_id = ?", sprintID)
	}

	if err := query.Preload("Assignee").Preload("Sprint").Find(&stories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user stories"})
		return
	}

	c.JSON(http.StatusOK, stories)
}

func GetUserStory(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	storyID := c.Param("storyId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var story models.UserStory
	if err := database.DB.Preload("Assignee").Preload("Sprint").First(&story, "id = ? AND project_id = ?", storyID, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User story not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, story)
}

func UpdateUserStory(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	storyID := c.Param("storyId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var req UpdateUserStoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var story models.UserStory
	if err := database.DB.First(&story, "id = ? AND project_id = ?", storyID, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User story not found"})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Acceptance != nil {
		updates["acceptance"] = req.Acceptance
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.StoryPoints != nil {
		updates["story_points"] = req.StoryPoints
	}
	if req.SprintID != nil {
		// If SprintID is empty string, maybe clear sprint?
		// Assuming UUID validation handles empty vs non-existent.
		// Check valid sprint
		var count int64
		database.DB.Model(&models.Sprint{}).Where("id = ? AND project_id = ?", *req.SprintID, projectID).Count(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
			return
		}
		updates["sprint_id"] = req.SprintID
	}
	if req.AssigneeID != nil {
		// Validate assignee
		if !isProjectMember(*req.AssigneeID, projectID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Assignee must be a project member"})
			return
		}
		updates["assignee_id"] = req.AssigneeID
	}

	if req.Status != nil {
		newStatus := *req.Status
		updates["status"] = newStatus
		if newStatus == "DONE" && story.Status != "DONE" {
			now := time.Now()
			updates["completed_at"] = &now
		} else if newStatus != "DONE" && story.Status == "DONE" {
			updates["completed_at"] = nil
		}
	}

	if err := database.DB.Model(&story).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user story"})
		return
	}

	// Fetch updated
	database.DB.Preload("Assignee").Preload("Sprint").First(&story, "id = ?", storyID)

	c.JSON(http.StatusOK, story)
}

func DeleteUserStory(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	storyID := c.Param("storyId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	if err := database.DB.Delete(&models.UserStory{}, "id = ? AND project_id = ?", storyID, projectID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user story"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User story deleted"})
}
