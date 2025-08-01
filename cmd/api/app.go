package api

import (
	"fmt"

	routes "github.com/crane-cloud/mira-new/cmd/api/routes"
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/open-ug/conveyor/pkg/client"
)

func StartServer(port string) {
	// Load environment variables from .env file

	// Initialize the client
	cl := client.NewClient()
	if cl == nil {
		panic("Failed to create client")
	}

	app := fiber.New(fiber.Config{
		AppName:     "MIRA API Server",
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Enable CORS
	app.Use(cors.New())

	app.Get("/api/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/api/echo", func(c *fiber.Ctx) error {
		var body map[string]interface{}
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid JSON")
		}
		return c.JSON(body)
	})

	app.Static("/uploads", "./uploads")

	routes.ImageRoutes(app, cl)
	routes.GitOAuthRoutes(app)
	routes.GitUserRoutes(app)

	app.Listen(":" + port)
	fmt.Println("Server started on port:", port)
}
