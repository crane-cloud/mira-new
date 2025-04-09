package routes

import (
	"conveyor.cloud.cranom.tech/pkg/client"
	handlers "github.com/crane-cloud/mira-new/cmd/api/handlers"
	"github.com/gofiber/fiber/v2"
)

func ImageRoutes(app *fiber.App, cl *client.Client) {
	imageHandler := handlers.NewImageHandler(cl)
	imagePrefix := app.Group("/images")

	imagePrefix.Post("/containerize", imageHandler.GenerateImage)

}
