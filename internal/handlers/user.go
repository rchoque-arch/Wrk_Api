package handlers

import (
	"net/http"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"Wrk_Api/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateUser(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	// Hash password if provided
	if user.Password != "" {
		hashed, err := utils.HashPassword(user.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
			return
		}
		user.Password = hashed
	}

	if result := database.DB.Create(&user); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func GetUsers(c *gin.Context) {
	var users []models.User
	if result := database.DB.Find(&users); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}
