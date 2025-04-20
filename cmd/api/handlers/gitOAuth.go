package handlers

import (
	"context"
	"github.com/crane-cloud/mira-new/cmd/config"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/oauth2"
	"log"
)

func GitHubLogin(c *fiber.Ctx) error {
	oauthConf := config.GitHubOAuthConfig()
	url := oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

func GitLabLogin(c *fiber.Ctx) error {
	oauthConf := config.GitLabOAuthConfig()
	url := oauthConf.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	return c.Redirect(url)
}

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
