// Package routes provides routes functionality.
package routes

import (
	"Wrk_Api/internal/handlers"
	"Wrk_Api/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes executes the SetupRoutes operation.
func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// Auth Routes (Public)
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

		// Project Routes (Protected)
		projects := api.Group("/projects")
		projects.Use(middleware.AuthMiddleware())
		{
			projects.POST("/", handlers.CreateProject)
			projects.GET("/", handlers.GetProjects)
			projects.GET("/:id", handlers.GetProject)
			projects.PUT("/:id", handlers.UpdateProject)
			projects.DELETE("/:id", handlers.DeleteProject)

			// Sprint Routes (Nested under Projects)
			sprints := projects.Group("/:projectId/sprints")
			{
				sprints.POST("/", handlers.CreateSprint)
				sprints.GET("/", handlers.GetSprints)
				sprints.GET("/:sprintId", handlers.GetSprint)
				sprints.PUT("/:sprintId", handlers.UpdateSprint)
				sprints.DELETE("/:sprintId", handlers.DeleteSprint)

				// Retrospective Routes (Nested under Sprints)
				retros := sprints.Group("/:sprintId/retrospectives")
				{
					retros.POST("/", handlers.CreateRetrospectiveItem)
					retros.GET("/", handlers.GetRetrospectiveItems)
					retros.PUT("/:itemId", handlers.UpdateRetrospectiveItem)
					retros.DELETE("/:itemId", handlers.DeleteRetrospectiveItem)
				}
			}

			// User Story Routes (Nested under Projects)
			stories := projects.Group("/:projectId/stories")
			{
				stories.POST("/", handlers.CreateUserStory)
				stories.GET("/", handlers.GetUserStories)
				stories.GET("/:storyId", handlers.GetUserStory)
				stories.PUT("/:storyId", handlers.UpdateUserStory)
				stories.DELETE("/:storyId", handlers.DeleteUserStory)
			}

			// Task Routes (Nested under Projects)
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
		}

		// WebSocket Route
		api.GET("/ws", middleware.AuthMiddleware(), handlers.ServeWs)
	}
}
