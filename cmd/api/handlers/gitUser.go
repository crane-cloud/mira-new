package handlers

import (
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
)

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
