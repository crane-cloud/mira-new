package handlers

import (
	"context"

	"conveyor.cloud.cranom.tech/pkg/client"
	cTypes "conveyor.cloud.cranom.tech/pkg/types"
	types "github.com/crane-cloud/mira-new/cmd/api/types"
	"github.com/gofiber/fiber/v2"
)

type ImageHandler struct {
	client *client.Client
}

func NewImageHandler(cl *client.Client) *ImageHandler {
	if cl == nil {
		cl = client.NewClient()
	}
	return &ImageHandler{
		client: cl,
	}
}

func (h *ImageHandler) GenerateImage(c *fiber.Ctx) error {

	var payload types.ImageRoutePayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request payload",
		})
	}

	var app cTypes.Application

	app.Name = payload.Name
	app.Spec.Source.Type = payload.SourceType
	if payload.SourceType == "git" {
		app.Spec.Source.GitRepo.URL = payload.SourceURL
	}

	_, err := h.client.CreateApplication(context.Background(), &app)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create application",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Image generation started",
		"data": fiber.Map{
			"name":        payload.Name,
			"source_type": payload.SourceType,
			"source_url":  payload.SourceURL,
		},
	})

}
