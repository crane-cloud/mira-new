package routes

import (
	"github.com/crane-cloud/mira-new/cmd/api/handlers"
	"github.com/gofiber/fiber/v2"
)

func GitOAuthRoutes(app *fiber.App) {
	// GitHub OAuth routes
	app.Get("/api/auth/github/login", handlers.GitHubLogin)
	app.Get("/api/auth/github/callback", handlers.GitHubCallback)

	// GitLab OAuth routes
	app.Get("/api/auth/gitlab/login", handlers.GitLabLogin)
	app.Get("/api/auth/gitlab/callback", handlers.GitLabCallback)
}
