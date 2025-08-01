package handlers

import (
	"context"
	"log"

	_ "mira/cmd/api/models"
	"mira/cmd/config"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
)

// GitHubLogin initiates GitHub OAuth login flow
// @Summary GitHub OAuth login
// @Description Redirects to GitHub OAuth authorization page
// @Tags auth
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to GitHub OAuth"
// @Router /auth/github/login [get]
func GitHubLogin(c *fiber.Ctx) error {
	oauthConf := config.GitHubOAuthConfig()
	url := oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

// GitLabLogin initiates GitLab OAuth login flow
// @Summary GitLab OAuth login
// @Description Redirects to GitLab OAuth authorization page
// @Tags auth
// @Accept json
// @Produce json
// @Success 302 {string} string "Redirect to GitLab OAuth"
// @Router /auth/gitlab/login [get]
func GitLabLogin(c *fiber.Ctx) error {
	oauthConf := config.GitLabOAuthConfig()
	url := oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

// GitHubCallback handles GitHub OAuth callback
// @Summary GitHub OAuth callback
// @Description Handles the callback from GitHub OAuth and sets authentication cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "OAuth authorization code"
// @Success 302 {string} string "Redirect to application"
// @Failure 400 {object} models.ErrorResponse "Missing authorization code"
// @Failure 500 {object} models.ErrorResponse "OAuth exchange failed"
// @Router /auth/github/callback [get]
func GitHubCallback(c *fiber.Ctx) error {
	oauthConf := config.GitHubOAuthConfig()
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing code"})
	}

	token, err := oauthConf.Exchange(context.Background(), code)
	if err != nil {
		log.Println("OAuth exchange error:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "OAuth exchange failed"})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    token.AccessToken,
		HTTPOnly: true,
		SameSite: "Lax",
	})

	// TODO: redirect to frontend
	return c.Redirect("")
}

// GitLabCallback handles GitLab OAuth callback
// @Summary GitLab OAuth callback
// @Description Handles the callback from GitLab OAuth and sets authentication cookie
// @Tags auth
// @Accept json
// @Produce json
// @Param code query string true "OAuth authorization code"
// @Success 302 {string} string "Redirect to application"
// @Failure 400 {object} models.ErrorResponse "Missing authorization code"
// @Failure 500 {object} models.ErrorResponse "OAuth exchange failed"
// @Router /auth/gitlab/callback [get]
func GitLabCallback(c *fiber.Ctx) error {
	oauthConf := config.GitLabOAuthConfig()
	code := c.Query("code")
	if code == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing code")
	}

	token, err := oauthConf.Exchange(context.Background(), code)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to exchange token")
	}

	// Set token in secure cookie
	c.Cookie(&fiber.Cookie{
		Name:     "gitlab_access_token",
		Value:    token.AccessToken,
		HTTPOnly: true,
		SameSite: "Lax",
	})

	// TODO: redirect to frontend
	return c.Redirect("")
}
