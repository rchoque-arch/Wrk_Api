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

type CreateSprintRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description *string   `json:"description"`
	StartDate   time.Time `json:"startDate" binding:"required"`
	EndDate     time.Time `json:"endDate" binding:"required"`
}

type UpdateSprintRequest struct {
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	Status      string     `json:"status"` // PLANNING, ACTIVE, COMPLETED
}

// Helper to check if user is a member of the project
func isProjectMember(userId string, projectId string) bool {
	var count int64
	// In GORM, Count counts the number of rows.
	// We need to query the project_members table
	err := database.DB.Model(&models.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectId, userId).
		Count(&count).Error

	if err != nil {
		return false
	}
	return count > 0
}

func CreateSprint(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)
	projectId := c.Param("projectId")

	// Validate Project Access
	if !isProjectMember(userId, projectId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var req CreateSprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.EndDate.Before(req.StartDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	sprint := models.Sprint{
		ID:          uuid.NewString(),
		ProjectID:   projectId,
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      "PLANNING",
	}

	if err := database.DB.Create(&sprint).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sprint"})
		return
	}

	c.JSON(http.StatusCreated, sprint)
}

func GetSprints(c *gin.Context) {
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

	var sprints []models.Sprint
	if err := database.DB.Where("project_id = ?", projectId).Find(&sprints).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sprints"})
		return
	}

	c.JSON(http.StatusOK, sprints)
}

func GetSprint(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)
	projectId := c.Param("projectId")
	sprintId := c.Param("sprintId")

	if !isProjectMember(userId, projectId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var sprint models.Sprint
	if err := database.DB.First(&sprint, "id = ? AND project_id = ?", sprintId, projectId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, sprint)
}

func UpdateSprint(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)
	projectId := c.Param("projectId")
	sprintId := c.Param("sprintId")

	if !isProjectMember(userId, projectId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var req UpdateSprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var sprint models.Sprint
	if err := database.DB.First(&sprint, "id = ? AND project_id = ?", sprintId, projectId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}

	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.StartDate != nil {
		updates["start_date"] = *req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = *req.EndDate
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}

	// Validate dates logic
	start := sprint.StartDate
	end := sprint.EndDate

	if val, ok := updates["start_date"].(time.Time); ok {
		start = val
	}
	if val, ok := updates["end_date"].(time.Time); ok {
		end = val
	}

	if end.Before(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End date must be after start date"})
		return
	}

	if err := database.DB.Model(&sprint).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sprint"})
		return
	}

	// Refresh sprint data
	database.DB.First(&sprint, "id = ?", sprintId)

	c.JSON(http.StatusOK, sprint)
}

func DeleteSprint(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)
	projectId := c.Param("projectId")
	sprintId := c.Param("sprintId")

	if !isProjectMember(userId, projectId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	// Ideally strictly check for roles, but skipping for now as per minimal viable plan

	result := database.DB.Delete(&models.Sprint{}, "id = ? AND project_id = ?", sprintId, projectId)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sprint"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Sprint deleted"})
}
