package handlers

import (
	"net/http"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"Wrk_Api/internal/realtime"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CreateChatRequest struct {
	ProjectID *string  `json:"projectId"`
	Type      string   `json:"type" binding:"required"` // PROJECT, DIRECT
	UserIDs   []string `json:"userIds"`                 // For DIRECT, list of other participants
	Title     *string  `json:"title"`
}

type SendMessageRequest struct {
	Content string `json:"content" binding:"required"`
}

// CreateChat creates a new chat room
func CreateChat(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)

	var req CreateChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	chatID := uuid.NewString()
	chat := models.Chat{
		ID:        chatID,
		Type:      req.Type,
		Title:     req.Title,
		ProjectID: req.ProjectID,
	}

	// Transaction to create chat and participants
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&chat).Error; err != nil {
			return err
		}

		// Add creator
		participants := []models.ChatParticipant{
			{ChatID: chatID, UserID: userID},
		}

		// Add other users
		if req.Type == "DIRECT" {
			for _, uid := range req.UserIDs {
				if uid != userID { // Avoid duplicate
					participants = append(participants, models.ChatParticipant{ChatID: chatID, UserID: uid})
				}
			}
		} else if req.Type == "PROJECT" && req.ProjectID != nil {
			// Optionally auto-add all project members? Or let them join?
			// For simplicity, let's assume project chats are open but we track participants for notification.
			// Or we can just add the creator. Let's stick to explicit participants for now.
		}

		if len(participants) > 0 {
			if err := tx.Create(&participants).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create chat"})
		return
	}

	c.JSON(http.StatusCreated, chat)
}

// GetUserChats returns chats the user is participating in
func GetUserChats(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)

	var chats []models.Chat
	// Join with participants
	err := database.DB.Distinct("chats.*").
		Joins("JOIN chat_participants ON chat_participants.chat_id = chats.id").
		Where("chat_participants.user_id = ?", userID).
		Preload("Participants").
		Preload("Participants.User"). // Load user details
		Find(&chats).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch chats"})
		return
	}

	c.JSON(http.StatusOK, chats)
}

// SendMessage adds a message to a chat
func SendMessage(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	chatID := c.Param("chatId")


	// Validate user is participant
	var count int64
	database.DB.Model(&models.ChatParticipant{}).Where("chat_id = ? AND user_id = ?", chatID, userID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to chat"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	message := models.Message{
		ID:      uuid.NewString(),
		ChatID:  chatID,
		UserID:  userID,
		Content: req.Content,
	}

	if err := database.DB.Create(&message).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Broadcast Event
	realtime.GlobalHub.BroadcastEvent(chatID, "MESSAGE_SENT", message)

	c.JSON(http.StatusCreated, message)
}

// GetMessages returns history for a chat
func GetMessages(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	chatID := c.Param("chatId")


	// Validate access
	var count int64
	database.DB.Model(&models.ChatParticipant{}).Where("chat_id = ? AND user_id = ?", chatID, userID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to chat"})
		return
	}

	var messages []models.Message
	if err := database.DB.Preload("User").Where("chat_id = ?", chatID).Order("created_at asc").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}
