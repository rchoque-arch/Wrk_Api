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

type AddMemberRequest struct {
	Email string `json:"email" binding:"required,email"`
	Role  string `json:"role"` // MEMBER, ADMIN (within project context, defaults to MEMBER)
}

func AddMember(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)
	projectId := c.Param("projectId")

	// Validate Access: Only OWNER can add members (for simplicity)
	// Or we could query the project role of the requester
	var requesterMember models.ProjectMember
	if err := database.DB.First(&requesterMember, "project_id = ? AND user_id = ?", projectId, userId).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}
	if requesterMember.Role != "OWNER" {
		// Maybe ADMIN role too? Sticking to OWNER for strictness initially.
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can add members"})
		return
	}

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find User by Email
	var userToAdd models.User
	if err := database.DB.First(&userToAdd, "email = ?", req.Email).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User with this email not found"})
		return
	}

	// Check if already member
	var count int64
	database.DB.Model(&models.ProjectMember{}).
		Where("project_id = ? AND user_id = ?", projectId, userToAdd.ID).
		Count(&count)

	if count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member of this project"})
		return
	}

	role := "MEMBER"
	if req.Role != "" {
		role = req.Role
	}

	member := models.ProjectMember{
		ID:        uuid.NewString(),
		ProjectID: projectId,
		UserID:    userToAdd.ID,
		Role:      role,
		JoinedAt:  time.Now(),
	}

	if err := database.DB.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add member"})
		return
	}

	// Return Member with User details
	database.DB.Preload("User").First(&member, "id = ?", member.ID)

	c.JSON(http.StatusCreated, member)
}

func GetMembers(c *gin.Context) {
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

	var members []models.ProjectMember
	if err := database.DB.Preload("User").Where("project_id = ?", projectId).Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}

	c.JSON(http.StatusOK, members)
}

func RemoveMember(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)
	projectId := c.Param("projectId")
	memberId := c.Param("memberId") // This is the ID of the ProjectMember record, or UserID? usually MemberID for REST consistency

	// Check Requester Role
	var requesterMember models.ProjectMember
	if err := database.DB.First(&requesterMember, "project_id = ? AND user_id = ?", projectId, userId).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}
	if requesterMember.Role != "OWNER" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only project owner can remove members"})
		return
	}

	var memberToRemove models.ProjectMember
	if err := database.DB.First(&memberToRemove, "id = ? AND project_id = ?", memberId, projectId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		}
		return
	}

	// Prevent removing self (Owner) via this endpoint?
	if memberToRemove.UserID == userId {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove yourself. Delete the project instead."})
		return
	}

	if err := database.DB.Delete(&memberToRemove).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove member"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Member removed"})
}
