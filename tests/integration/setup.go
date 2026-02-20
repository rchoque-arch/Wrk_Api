package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"Wrk_Api/internal/database"
	"Wrk_Api/internal/handlers"
	"Wrk_Api/internal/middleware"
	"Wrk_Api/internal/models"
	"Wrk_Api/internal/realtime"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func SetupRouter() *gin.Engine {
	// Start GlobalHub for tests to avoid deadlocks on broadcast
	go realtime.GlobalHub.Run()

	gin.SetMode(gin.TestMode)
	r := gin.Default()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
		}

		// User Routes (Protected)
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware())
		{
			users.GET("/me", handlers.GetMe)
			users.PUT("/me", handlers.UpdateMe)
			users.POST("/", handlers.CreateUser)
			users.GET("/", handlers.GetUsers)
		}

		projects := api.Group("/projects")
		projects.Use(middleware.AuthMiddleware())
		{
			projects.POST("/", handlers.CreateProject)
			projects.GET("/", handlers.GetProjects)

			// We cannot mix :id and :projectId at the same level in Gin.
			// Handlers use c.Param("id") for Project actions and c.Param("projectId") for nested actions.
			// This is a design flaw in the handler implementation we inherited/wrote.
			// Best fix: Update handlers/project.go to use "projectId" instead of "id".
			// But for now in setup.go, we can just use one param name and rely on Gin or fix the handlers.
			// Let's assume we fixed handlers to look for "projectId" if "id" is missing or just update handlers.
			// Updating handlers is cleaner but out of scope for just fixing tests? No, I can update handlers.

			// Let's use :projectId for everything here to satisfy Gin's constraints.
			// And I will update project.go handlers to read "projectId" OR "id".
			projects.GET("/:projectId", handlers.GetProject)
			projects.PUT("/:projectId", handlers.UpdateProject)
			projects.DELETE("/:projectId", handlers.DeleteProject)

			// Use :projectId to match what the handler expects
			sprints := projects.Group("/:projectId/sprints")
			{
				sprints.POST("/", handlers.CreateSprint)
				sprints.GET("/", handlers.GetSprints)
				sprints.GET("/:sprintId", handlers.GetSprint)
				sprints.PUT("/:sprintId", handlers.UpdateSprint)
				sprints.DELETE("/:sprintId", handlers.DeleteSprint)
			}

			stories := projects.Group("/:projectId/stories")
			{
				stories.POST("/", handlers.CreateUserStory)
				stories.GET("/", handlers.GetUserStories)
				stories.GET("/:storyId", handlers.GetUserStory)
				stories.PUT("/:storyId", handlers.UpdateUserStory)
				stories.DELETE("/:storyId", handlers.DeleteUserStory)
			}

			tasks := projects.Group("/:projectId/tasks")
			{
				tasks.POST("/", handlers.CreateTask)
				tasks.GET("/", handlers.GetTasks)
				tasks.GET("/:taskId", handlers.GetTask)
				tasks.PUT("/:taskId", handlers.UpdateTask)
				tasks.DELETE("/:taskId", handlers.DeleteTask)
			}

			// Rubric Routes
			rubrics := projects.Group("/:projectId/rubrics")
			{
				rubrics.POST("/", handlers.CreateRubric)
				rubrics.GET("/", handlers.GetRubrics)
				rubrics.GET("/:rubricId", handlers.GetRubric)
				rubrics.DELETE("/:rubricId", handlers.DeleteRubric)
			}

			// Evaluation Routes
			evaluations := projects.Group("/:projectId/evaluations")
			{
				evaluations.POST("/", handlers.CreateEvaluation)
				evaluations.GET("/", handlers.GetEvaluations)
			}

			// Document Routes
			docs := projects.Group("/:projectId/documents")
			{
				docs.POST("/", handlers.UploadDocument)
				docs.GET("/", handlers.GetDocuments)
				docs.DELETE("/:docId", handlers.DeleteDocument)
				docs.GET("/:docId/download", handlers.DownloadDocument)
			}

			// Metrics Route
			projects.GET("/:projectId/metrics", handlers.GetProjectMetrics)

			// Member Routes
			members := projects.Group("/:projectId/members")
			{
				members.POST("/", handlers.AddMember)
				members.GET("/", handlers.GetMembers)
				members.DELETE("/:memberId", handlers.RemoveMember)
			}

			// Retrospective Routes (Nested under Sprints)
			// Need to match route structure: /projects/:projectId/sprints/:sprintId/retrospectives
			// Currently defined inside 'sprints' group but let's verify if that group is using :projectId
			// Yes: sprints := projects.Group("/:projectId/sprints")
			// So retros are: /projects/:projectId/sprints/:sprintId/retrospectives. Correct.
			// But wait, the 'sprints' group above didn't include Retrospectives inside its block in this file.
			// I need to fix that or add it separately.
			// Adding it separately is tricky because of the nested group context.
			// Let's rely on the block above? No, the block above for sprints only has CRUD for sprints.
			// I need to patch the sprints block to include retrospectives or add a new group.
			// Adding a new group:
			sprintRetros := projects.Group("/:projectId/sprints/:sprintId/retrospectives")
			{
				sprintRetros.POST("/", handlers.CreateRetrospectiveItem)
				sprintRetros.GET("/", handlers.GetRetrospectiveItems)
				sprintRetros.PUT("/:itemId", handlers.UpdateRetrospectiveItem)
				sprintRetros.DELETE("/:itemId", handlers.DeleteRetrospectiveItem)
			}
		}

		// Notification Routes (Protected)
		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware())
		{
			notifications.GET("/", handlers.GetNotifications)
			notifications.PUT("/:id/read", handlers.MarkNotificationRead)
		}

		// Chat Routes (Protected)
		chats := api.Group("/chats")
		chats.Use(middleware.AuthMiddleware())
		{
			chats.POST("/", handlers.CreateChat)
			chats.GET("/", handlers.GetUserChats)
			chats.POST("/:chatId/messages", handlers.SendMessage)
			chats.GET("/:chatId/messages", handlers.GetMessages)
		}

		// WebSocket Route
		api.GET("/ws", middleware.AuthMiddleware(), handlers.ServeWs)
	}
	return r
}

func SetupTestDB() {
	var err error
	database.DB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	database.DB.AutoMigrate(
		&models.User{},
		&models.Project{},
		&models.ProjectMember{},
		&models.Sprint{},
		&models.UserStory{},
		&models.Task{},
		&models.Rubric{},
		&models.Criteria{},
		&models.Evaluation{},
		&models.EvaluationCriteria{},
		&models.RetrospectiveItem{},
		&models.Notification{},
		&models.Chat{},
		&models.ChatParticipant{},
		&models.Message{},
		&models.Document{},
	)
}

func GetAuthToken(r *gin.Engine, email, name string) (string, string) {
	registerReq := handlers.RegisterRequest{
		Email:    email,
		Name:     name,
		Password: "password123",
	}
	jsonValue, _ := json.Marshal(registerReq)
	req, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// If already exists, login instead
	if w.Code == http.StatusConflict {
		loginReq := handlers.LoginRequest{
			Email:    email,
			Password: "password123",
		}
		jsonValue, _ = json.Marshal(loginReq)
		req, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(jsonValue))
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}

	var response handlers.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	return response.Token, response.User.ID
}
