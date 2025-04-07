package api

import (
	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func StartServer(port string) {
	app := fiber.New(fiber.Config{
		AppName:     "MIRA API Server",
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Enable CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

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

	app.Listen(":" + port)
}
