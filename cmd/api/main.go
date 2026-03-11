// Package main provides the API server entry point.
package main

import (
	"log"
	"os"
	"regexp"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/realtime"
	"Wrk_Api/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using defaults")
	}

	database.InitDB()

	// Start WebSocket Hub
	go realtime.GlobalHub.Run()

	r := gin.Default()

	routes.SetupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	validPort := regexp.MustCompile(`^\d+$`)
	if !validPort.MatchString(port) {
		port = "8080"
	}

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	//nolint:gosec // Port is validated with regex
	log.Printf("Starting server on port %q", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
