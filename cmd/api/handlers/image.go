package handlers

import (
	"github.com/gofiber/fiber/v2"
)

type ImageHandler struct {
}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

func (h *ImageHandler) GenerateImage(c *fiber.Ctx) error {

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"id": 1,
	})
}
