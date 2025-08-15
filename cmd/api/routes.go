package api

import (
	handlers "mira/cmd/api/handlers"
	"mira/cmd/api/services"
	common "mira/cmd/common"
	"mira/cmd/config"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, natsClient *common.NATSClient, mongoConfig *config.MongoDBConfig) {
	// Initialize MongoDB service
	var mongoService *services.MongoLogService
	if mongoConfig != nil && mongoConfig.Client != nil {
		mongoService = services.NewMongoLogService(mongoConfig)
	}

	// Setup all route groups
	setupImageRoutes(app, natsClient)
	setupLogRoutes(app, natsClient, mongoService)
	setupGitUserRoutes(app)
	setupGitOAuthRoutes(app)
}

// setupImageRoutes configures image containerization routes
func setupImageRoutes(app *fiber.App, natsClient *common.NATSClient) {
	imageHandler := handlers.NewImageHandler(natsClient)
	if imageHandler == nil {
		panic("Failed to create image handler")
	}

	imagePrefix := app.Group("/api/images")

	imagePrefix.Post("/containerize", imageHandler.GenerateImage)
	imagePrefix.Post("/detect", handlers.DetectFramework)
}

// setupLogRoutes configures WebSocket log streaming routes
func setupLogRoutes(app *fiber.App, natsClient *common.NATSClient, mongoService *services.MongoLogService) {
	logHandler := handlers.NewLogHandler(natsClient, mongoService)

	// WebSocket endpoint for streaming logs
	app.Get("/api/logs/:buildId", logHandler.WebSocketUpgrade)
	// app.Get("/api/logs/:buildId/history", logHandler.GetBuildLogs)

	// MongoDB-based endpoints
	app.Get("/api/logs", logHandler.GetBuildLogsFromMongoDB)
	app.Get("/api/logs/stats", logHandler.GetLogStats)
	app.Get("/api/builds", logHandler.GetBuilds)
}

// setupGitUserRoutes configures Git user repository routes
func setupGitUserRoutes(app *fiber.App) {
	// GitHub user routes
	app.Get("/api/user/github/repos", handlers.GetGithubRepositories)

	// GitLab user routes
	app.Get("/api/user/gitlab/repos", handlers.GetGitlabRepositories)
}

// setupGitOAuthRoutes configures Git OAuth authentication routes
func setupGitOAuthRoutes(app *fiber.App) {
	// GitHub OAuth routes
	app.Get("/api/auth/github/login", handlers.GitHubLogin)
	app.Get("/api/auth/github/callback", handlers.GitHubCallback)

	// GitLab OAuth routes
	app.Get("/api/auth/gitlab/login", handlers.GitLabLogin)
	app.Get("/api/auth/gitlab/callback", handlers.GitLabCallback)
}
