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

type CreateProjectRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description *string    `json:"description"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
}

type UpdateProjectRequest struct {
	Name        string     `json:"name"`
	Description *string    `json:"description"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
}

func CreateProject(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectId := uuid.NewString()

	project := models.Project{
		ID:          projectId,
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		OwnerID:     userId,
		Status:      "ACTIVE",
	}

	// Transaction: Create project AND add owner as a member
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&project).Error; err != nil {
			return err
		}

		member := models.ProjectMember{
			ID:        uuid.NewString(),
			ProjectID: projectId,
			UserID:    userId,
			Role:      "OWNER",
		}

		if err := tx.Create(&member).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func GetProjects(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)

	var projects []models.Project
	// Correct query: find projects where user is a member
	// JOIN project_members ON project_members.project_id = projects.id AND project_members.user_id = ?
	err := database.DB.
		Joins("JOIN project_members ON project_members.project_id = projects.id").
		Where("project_members.user_id = ?", userId).
		Find(&projects).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

func GetProject(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)

	projectId := c.Param("id")
	if projectId == "" {
		projectId = c.Param("projectId")
	}

	var project models.Project
	// Preload members to check access
	err := database.DB.Preload("Members").First(&project, "id = ?", projectId).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Access Control: Must be a member or owner
	isMember := false
	if project.OwnerID == userId {
		isMember = true
	} else {
		for _, member := range project.Members {
			if member.UserID == userId {
				isMember = true
				break
			}
		}
	}

	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func UpdateProject(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)

	projectId := c.Param("id")
	if projectId == "" {
		projectId = c.Param("projectId")
	}
	var req UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var project models.Project
	if err := database.DB.First(&project, "id = ?", projectId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Only owner can update for now
	if project.OwnerID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can update the project"})
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != nil {
		updates["description"] = req.Description
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.StartDate != nil {
		updates["start_date"] = req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = req.EndDate
	}

	if err := database.DB.Model(&project).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	c.JSON(http.StatusOK, project)
}

func DeleteProject(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)

	projectId := c.Param("id")
	if projectId == "" {
		projectId = c.Param("projectId")
	}

	var project models.Project
	if err := database.DB.First(&project, "id = ?", projectId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	if project.OwnerID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the owner can delete the project"})
		return
	}

	if err := database.DB.Delete(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted"})
}
