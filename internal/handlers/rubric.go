package handlers

import (
	"net/http"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CreateCriteriaRequest represents the CreateCriteriaRequest structure.
type CreateCriteriaRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	MaxScore    int     `json:"maxScore"`
	Weight      int     `json:"weight"`
}

// CreateRubricRequest represents the CreateRubricRequest structure.
type CreateRubricRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description *string                 `json:"description"`
	Criteria    []CreateCriteriaRequest `json:"criteria" binding:"required,dive"`
}

// CreateRubric executes the CreateRubric operation.
func CreateRubric(c *gin.Context) {
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

	var req CreateRubricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rubricID := uuid.NewString()
	rubric := models.Rubric{
		ID:          rubricID,
		ProjectID:   &projectID,
		Name:        req.Name,
		Description: req.Description,
	}

	// Build criteria
	var criteriaList []models.Criteria
	for _, critReq := range req.Criteria {
		criteriaList = append(criteriaList, models.Criteria{
			ID:          uuid.NewString(),
			RubricID:    rubricID,
			Name:        critReq.Name,
			Description: critReq.Description,
			MaxScore:    critReq.MaxScore,
			Weight:      critReq.Weight,
		})
	}
	rubric.Criteria = criteriaList

	if err := database.DB.Create(&rubric).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create rubric"})
		return
	}

	c.JSON(http.StatusCreated, rubric)
}

// GetRubrics executes the GetRubrics operation.
func GetRubrics(c *gin.Context) {
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

	var rubrics []models.Rubric
	if err := database.DB.Preload("Criteria").Where("project_id = ?", projectID).Find(&rubrics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rubrics"})
		return
	}

	c.JSON(http.StatusOK, rubrics)
}

// GetRubric executes the GetRubric operation.
func GetRubric(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	rubricID := c.Param("rubricId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var rubric models.Rubric
	if err := database.DB.Preload("Criteria").
		First(&rubric, "id = ? AND project_id = ?", rubricID, projectID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Rubric not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, rubric)
}

// DeleteRubric executes the DeleteRubric operation.
func DeleteRubric(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Param("projectId")
	rubricID := c.Param("rubricId")

	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	if err := database.DB.Delete(&models.Rubric{}, "id = ? AND project_id = ?", rubricID, projectID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete rubric"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rubric deleted"})
}
