package handlers

import (
	"net/http"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateRetrospectiveItemRequest struct {
	Type    string `json:"type" binding:"required"` // GOOD, BAD, ACTION
	Content string `json:"content" binding:"required"`
}

type UpdateRetrospectiveItemRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

func CreateRetrospectiveItem(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	sprintID := c.Param("sprintId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	// Validate Sprint belongs to Project
	var sprint models.Sprint
	if err := database.DB.First(&sprint, "id = ? AND project_id = ?", sprintID, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sprint not found"})
		return
	}

	var req CreateRetrospectiveItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := models.RetrospectiveItem{
		ID:       uuid.NewString(),
		SprintID: sprintID,
		Type:     req.Type,
		Content:  req.Content,
		UserID:   userID,
	}

	if err := database.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create retrospective item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func GetRetrospectiveItems(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	sprintID := c.Param("sprintId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var items []models.RetrospectiveItem
	if err := database.DB.Preload("User").Where("sprint_id = ?", sprintID).Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch retrospective items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func UpdateRetrospectiveItem(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	// sprintID is in path but mostly needed for validation context if we enforced strict hierarchy checks
	itemID := c.Param("itemId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var item models.RetrospectiveItem
	if err := database.DB.First(&item, "id = ?", itemID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		return
	}

	// Only author can update? Or anyone in project? Usually author or scrum master.
	// Let's restrict to author for now.
	if item.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the author can update this item"})
		return
	}

	var req UpdateRetrospectiveItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}

	if err := database.DB.Model(&item).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func DeleteRetrospectiveItem(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	itemID := c.Param("itemId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var item models.RetrospectiveItem
	if err := database.DB.First(&item, "id = ?", itemID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
		}
		return
	}

	if item.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the author can delete this item"})
		return
	}

	if err := database.DB.Delete(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item deleted"})
}
