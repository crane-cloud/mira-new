package routes

import (
	handlers "github.com/crane-cloud/mira-new/cmd/api/handlers"
	"github.com/gofiber/fiber/v2"
	"github.com/open-ug/conveyor/pkg/client"
)

func ImageRoutes(app *fiber.App, cl *client.Client) {
	imageHandler := handlers.NewImageHandler(cl)
	imagePrefix := app.Group("/images")

	imagePrefix.Post("/containerize", imageHandler.GenerateImage)

}
