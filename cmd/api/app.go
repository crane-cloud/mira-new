// Package api provides the MIRA API server
package api

import (
	"fmt"

	common "mira/cmd/common"
	_ "mira/docs" // Import generated docs

	"github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// @title MIRA API
// @version 1.0
// @description Auto-containerization platform API for building and deploying applications
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@mira.dev
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:3000
// @BasePath /api
// @schemes http https
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization

func StartServer(port string) {
	// Load environment variables from .env file

	// Initialize NATS client
	natsClient, err := common.NewNATSClient()
	if err != nil {
		panic(fmt.Sprintf("Failed to create NATS client: %v", err))
	}
	defer natsClient.Close()

	app := fiber.New(fiber.Config{
		AppName:     "MIRA API Server",
		JSONEncoder: json.Marshal,
		JSONDecoder: json.Unmarshal,
	})

	// Enable CORS
	app.Use(cors.New())

	// Health check endpoint
	// @Summary Health check
	// @Description Returns the health status of the API
	// @Tags health
	// @Accept json
	// @Produce plain
	// @Success 200 {string} string "OK"
	// @Router /health [get]
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

	// Serve static files from public directory
	app.Static("/", "./public")

	// Swagger documentation
	app.Get("/api/docs/*", fiberSwagger.WrapHandler)

	// Setup all API routes
	SetupRoutes(app, natsClient)

	app.Listen(":" + port)
	fmt.Println("Server started on port:", port)
}
