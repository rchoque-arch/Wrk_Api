package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UploadDocument handles file uploads linked to a project
func UploadDocument(c *gin.Context) {
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

	// Retrieve file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}

	// Ensure uploads dir exists
	uploadPath := "uploads"
	if _, err := os.Stat(uploadPath); os.IsNotExist(err) {
		_ = os.Mkdir(uploadPath, 0o750)
	}

	// Generate safe filename
	ext := filepath.Ext(file.Filename)
	newFilename := fmt.Sprintf("%s_%d%s", uuid.NewString(), time.Now().Unix(), ext)
	dst := filepath.Join(uploadPath, newFilename)

	// Save file
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Get file info
	sizeKB := int(file.Size / 1024)

	// Versioning Logic
	version := 1
	var parentId *string
	if pid := c.Query("parentId"); pid != "" {
		var parentDoc models.Document
		if err := database.DB.First(&parentDoc, "id = ? AND project_id = ?", pid, projectId).Error; err == nil {
			// Found parent, increment version
			// Ideally check if parent is latest or find max version in chain.
			// Simplified: New upload becomes latest version of provided parent.
			// Better: Find the latest version linked to this parent to get next number.

			// If parent itself has a parent, use the root parent.
			rootId := parentDoc.ID
			if parentDoc.ParentID != nil {
				rootId = *parentDoc.ParentID
			}
			parentId = &rootId

			// Find max version
			var maxVer int
			_ = database.DB.Model(&models.Document{}).
				Where("id = ? OR parent_id = ?", rootId, rootId).
				Select("MAX(version)").
				Row().Scan(&maxVer)

			if maxVer == 0 { maxVer = parentDoc.Version } // Fallback
			version = maxVer + 1
		}
	}

	// Create DB Record
	doc := models.Document{
		ID:        uuid.NewString(),
		ProjectID: projectId,
		Name:      file.Filename,
		URL:       dst, // Relative path
		Type:      ext,
		Size:      &sizeKB,
		Version:   version,
		ParentID:  parentId,
	}

	if err := database.DB.Create(&doc).Error; err != nil {
		// Clean up file if DB fails
		_ = os.Remove(dst)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save document metadata"})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

func GetDocuments(c *gin.Context) {
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

	var docs []models.Document
	if err := database.DB.Where("project_id = ?", projectId).Find(&docs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch documents"})
		return
	}

	c.JSON(http.StatusOK, docs)
}

func DeleteDocument(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)
	projectId := c.Param("projectId")
	docId := c.Param("docId")

	if !isProjectMember(userId, projectId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var doc models.Document
	if err := database.DB.First(&doc, "id = ? AND project_id = ?", docId, projectId).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		}
		return
	}

	// Delete file from disk
	filePath := doc.URL
	_ = os.Remove(filePath) // Ignore error, maybe already gone

	if err := database.DB.Delete(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted"})
}

// DownloadDocument serves the file content
func DownloadDocument(c *gin.Context) {
	userIdStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userId := userIdStr.(string)
	projectId := c.Param("projectId")
	docId := c.Param("docId")

	if !isProjectMember(userId, projectId) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	var doc models.Document
	if err := database.DB.First(&doc, "id = ? AND project_id = ?", docId, projectId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.File(doc.URL)
}
