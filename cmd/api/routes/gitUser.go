package routes

import (
	"github.com/crane-cloud/mira-new/cmd/api/handlers"
	"github.com/gofiber/fiber/v2"
)

func GitUserRoutes(app *fiber.App) {
	// GitHub user routes
	app.Get("/api/user/github/repos", handlers.GetGithubRepositories)

	// GitLab user routes
	app.Get("/api/user/gitlab/repos", handlers.GetGitlabRepositories)
}
