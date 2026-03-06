package handlers

import (
	"net/http"

	"Wrk_Api/internal/realtime"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for simplicity in development
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWs handles websocket requests from the peer.
func ServeWs(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userID := userIDStr.(string)
	projectID := c.Query("projectId") // Client must specify which project they are viewing

	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID required"})
		return
	}

	// Validate membership
	if !isProjectMember(userID, projectID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to project"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := &realtime.Client{
		Hub:       realtime.GlobalHub,
		Conn:      conn,
		Send:      make(chan realtime.Message, 256),
		UserID:    userID,
		ProjectID: projectID,
	}

	client.Hub.Register() <- client

	// Start go routines
	go client.WritePump()
	go client.ReadPump()
}
