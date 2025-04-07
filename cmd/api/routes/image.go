package routes

import (
	handlers "github.com/crane-cloud/mira-new/cmd/api/handlers"
	"github.com/gofiber/fiber/v2"
)

func ImageRoutes(app *fiber.App) {
	imageHandler := handlers.NewImageHandler()
	imagePrefix := app.Group("/images")

	imagePrefix.Post("/containerize", imageHandler.GenerateImage)

}
