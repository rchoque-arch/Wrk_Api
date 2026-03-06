package handlers

import (
	"net/http"
	"time"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"Wrk_Api/internal/realtime"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description *string    `json:"description"`
	Priority    string     `json:"priority"` // MEDIUM
	Deadline    *time.Time `json:"deadline"`
	UserStoryID *string    `json:"userStoryId"`
	SprintID    *string    `json:"sprintId"`
	AssigneeID  *string    `json:"assigneeId"`
}

type UpdateTaskRequest struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Priority    *string    `json:"priority"`
	Status      *string    `json:"status"` // TODO, IN_PROGRESS, REVIEW, DONE
	Deadline    *time.Time `json:"deadline"`
	UserStoryID *string    `json:"userStoryId"`
	SprintID    *string    `json:"sprintId"`
	AssigneeID  *string    `json:"assigneeId"`
}

func CreateTask(c *gin.Context) {
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

	var req CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation
	if req.UserStoryID != nil {
		var count int64
		database.DB.Model(&models.UserStory{}).Where("id = ? AND project_id = ?", *req.UserStoryID, projectID).Count(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user story ID"})
			return
		}
	}

	if req.SprintID != nil {
		var count int64
		database.DB.Model(&models.Sprint{}).Where("id = ? AND project_id = ?", *req.SprintID, projectID).Count(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
			return
		}
	}

	if req.AssigneeID != nil {
		if !isProjectMember(*req.AssigneeID, projectID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Assignee must be a project member"})
			return
		}
	}

	task := models.Task{
		ID:          uuid.NewString(),
		ProjectID:   projectID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    "MEDIUM",
		Status:      "TODO",
		Deadline:    req.Deadline,
		UserStoryID: req.UserStoryID,
		SprintID:    req.SprintID,
		AssigneeID:  req.AssigneeID,
	}

	if req.Priority != "" {
		task.Priority = req.Priority
	}

	if err := database.DB.Create(&task).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	// Broadcast Event
	realtime.GlobalHub.BroadcastEvent(projectID, "TASK_CREATED", task)

	c.JSON(http.StatusCreated, task)
}

func GetTasks(c *gin.Context) {
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

	var tasks []models.Task
	query := database.DB.Where("project_id = ?", projectID)

	// Filtering
	if sprintID := c.Query("sprintId"); sprintID != "" {
		query = query.Where("sprint_id = ?", sprintID)
	}
	if userStoryId := c.Query("userStoryId"); userStoryId != "" {
		query = query.Where("user_story_id = ?", userStoryId)
	}
	if assigneeId := c.Query("assigneeId"); assigneeId != "" {
		query = query.Where("assignee_id = ?", assigneeId)
	}

	if err := query.Preload("Assignee").Preload("Sprint").Preload("UserStory").Find(&tasks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func GetTask(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var task models.Task
	if err := database.DB.Preload("Assignee").Preload("Sprint").Preload("UserStory").First(&task, "id = ? AND project_id = ?", taskID, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

func UpdateTask(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var req UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var task models.Task
	if err := database.DB.First(&task, "id = ? AND project_id = ?", taskID, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Priority != nil {
		updates["priority"] = *req.Priority
	}
	if req.Deadline != nil {
		updates["deadline"] = req.Deadline
	}

	// Validations for FKs
	if req.UserStoryID != nil {
		var count int64
		database.DB.Model(&models.UserStory{}).Where("id = ? AND project_id = ?", *req.UserStoryID, projectID).Count(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user story ID"})
			return
		}
		updates["user_story_id"] = req.UserStoryID
	}
	if req.SprintID != nil {
		var count int64
		database.DB.Model(&models.Sprint{}).Where("id = ? AND project_id = ?", *req.SprintID, projectID).Count(&count)
		if count == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid sprint ID"})
			return
		}
		updates["sprint_id"] = req.SprintID
	}
	if req.AssigneeID != nil {
		if !isProjectMember(*req.AssigneeID, projectID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Assignee must be a project member"})
			return
		}
		updates["assignee_id"] = req.AssigneeID
	}

	if req.Status != nil {
		newStatus := *req.Status
		updates["status"] = newStatus
		if newStatus == "DONE" && task.Status != "DONE" {
			now := time.Now()
			updates["completed_at"] = &now
		} else if newStatus != "DONE" && task.Status == "DONE" {
			updates["completed_at"] = nil // This sets it to NULL in DB if using GORM map updates correctly with pointer or sql.NullTime
			// Since completed_at is *time.Time, setting it to nil in map updates usually works if GORM is configured right.
			// However, in Go map[string]interface{}, nil values are ignored by GORM updates by default unless using Select or specific config.
			// Let's force it for now.
			updates["completed_at"] = nil
		}
	}

	// Using Updates with map
	if err := database.DB.Model(&task).Select("completed_at", "title", "description", "priority", "deadline", "user_story_id", "sprint_id", "assignee_id", "status").Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	database.DB.Preload("Assignee").Preload("Sprint").Preload("UserStory").First(&task, "id = ?", taskID)

	// Broadcast Event
	realtime.GlobalHub.BroadcastEvent(projectID, "TASK_UPDATED", task)

	c.JSON(http.StatusOK, task)
}

func DeleteTask(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	taskID := c.Param("taskId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	if err := database.DB.Delete(&models.Task{}, "id = ? AND project_id = ?", taskID, projectID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		return
	}

	// Broadcast Event
	realtime.GlobalHub.BroadcastEvent(projectID, "TASK_DELETED", gin.H{"id": taskID})

	c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
}
