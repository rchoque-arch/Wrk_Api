// Package database provides database functionality.
package database

import (
	"log"
	"os"

	"Wrk_Api/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB is the exported variable.
var DB *gorm.DB

// InitDB executes the InitDB operation.
func InitDB() {
	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "wrk_api.db"
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	log.Println("Database connection established")

	// Auto Migrate
	err = DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.Document{},
		&models.ProjectMember{},
		&models.Sprint{},
		&models.UserStory{},
		&models.Task{},
		&models.Rubric{},
		&models.Criteria{},
		&models.Evaluation{},
		&models.EvaluationCriteria{},
		&models.Chat{},
		&models.ChatParticipant{},
		&models.Message{},
		&models.Notification{},
		&models.RetrospectiveItem{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	log.Println("Database migration completed")
}
