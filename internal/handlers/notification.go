package handlers

import (
	"net/http"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Utility function to create a notification (internal use)
func CreateNotification(userID, title, message, notifType string) error {
	notification := models.Notification{
		ID:      uuid.NewString(),
		UserID:  userID,
		Title:   title,
		Message: message,
		Type:    notifType,
		Read:    false,
	}
	return database.DB.Create(&notification).Error
}

func GetNotifications(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)

	var notifications []models.Notification
	if err := database.DB.Where("user_id = ?", userID).Order("created_at desc").Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func MarkNotificationRead(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	notificationId := c.Param("id")

	// Ensure notification belongs to user
	result := database.DB.Model(&models.Notification{}).
		Where("id = ? AND user_id = ?", notificationId, userID).
		Update("read", true)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update notification"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}
