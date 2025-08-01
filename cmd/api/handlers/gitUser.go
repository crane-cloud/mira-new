package handlers

import (
	"io"
	"net/http"

	_ "mira/cmd/api/models"

	"github.com/gofiber/fiber/v2"
)

// GetGithubRepositories fetches user's GitHub repositories
// @Summary Get GitHub repositories
// @Description Fetches repositories for the authenticated GitHub user
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} []models.GitHubRepository "List of repositories"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - no access token"
// @Failure 500 {object} models.ErrorResponse "GitHub API request failed"
// @Router /user/github/repos [get]
// @Security ApiKeyAuth
func GetGithubRepositories(c *fiber.Ctx) error {
	accessToken := c.Cookies("access_token")
	if accessToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No access token",
		})
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://api.github.com/user/repos?per_page=100", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "GitHub request failed",
		})
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return c.Status(resp.StatusCode).Send(body)
}

// GetGitlabRepositories fetches user's GitLab repositories
// @Summary Get GitLab repositories
// @Description Fetches repositories for the authenticated GitLab user
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} []models.GitLabProject "List of repositories"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - no access token"
// @Failure 500 {object} models.ErrorResponse "GitLab API request failed"
// @Router /user/gitlab/repos [get]
// @Security ApiKeyAuth
func GetGitlabRepositories(c *fiber.Ctx) error {
	accessToken := c.Cookies("gitlab_access_token")
	if accessToken == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "No access token found",
		})
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", "https://gitlab.com/api/v4/projects?membership=true", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err,
		})
	}

	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return c.Status(resp.StatusCode).Send(body)
}
