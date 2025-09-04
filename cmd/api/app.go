// Package api provides the MIRA API server
package api

import (
	"encoding/json"
	"fmt"
	"log"

	"mira/cmd/api/services"
	common "mira/cmd/common"
	"mira/cmd/config"
	_ "mira/docs" // Import generated docs

	gojson "github.com/goccy/go-json"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/nats-io/nats.go"
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

	// Initialize MongoDB
	mongoConfig := config.NewMongoDBConfig()
	err = mongoConfig.Connect()
	if err != nil {
		log.Printf("Warning: Failed to connect to MongoDB: %v", err)
		log.Printf("MongoDB features will be disabled")
		mongoConfig = nil
	} else {
		defer mongoConfig.Disconnect()
	}

	app := fiber.New(fiber.Config{
		AppName:     "MIRA API Server",
		JSONEncoder: gojson.Marshal,
		JSONDecoder: gojson.Unmarshal,
	})

	// Enable CORS with proper configuration
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "*",
		AllowMethods:     "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization,X-Requested-With",
		AllowCredentials: false,
		ExposeHeaders:    "Content-Length",
		MaxAge:           12 * 3600, // 12 hours
	}))

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
	app.Get("/docs/*", fiberSwagger.WrapHandler)

	// Setup all API routes
	SetupRoutes(app, natsClient, mongoConfig)

	// Start MongoDB log subscriber if MongoDB is available
	if mongoConfig != nil && mongoConfig.Client != nil {
		mongoService := services.NewMongoLogService(mongoConfig)
		startMongoDBLogSubscriber(natsClient, mongoService)
		startMongoDBBuildStatusSubscriber(natsClient, mongoService)
	}

	app.Listen(":" + port)
	fmt.Println("Server started on port:", port)
}

// startMongoDBLogSubscriber starts listening to NATS logs and saving them to MongoDB
func startMongoDBLogSubscriber(natsClient *common.NATSClient, mongoService *services.MongoLogService) {
	// Subscribe to all log subjects
	subject := "mira.logs.*"
	_, err := natsClient.GetConnection().Subscribe(subject, func(msg *nats.Msg) {
		var logMsg common.LogMessage
		if err := json.Unmarshal(msg.Data, &logMsg); err != nil {
			log.Printf("Failed to unmarshal log message: %v", err)
			return
		}

		// Save to MongoDB
		err := mongoService.SaveLog(
			logMsg.BuildID,
			logMsg.Level,
			logMsg.Message,
			logMsg.Step,
			logMsg.Timestamp,
		)
		if err != nil {
			log.Printf("Failed to save log to MongoDB: %v", err)
		} else {
			log.Printf("✅ Saved log to MongoDB for build %s: %s", logMsg.BuildID, logMsg.Message)
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe to logs: %v", err)
	} else {
		log.Printf("Started MongoDB log subscriber on subject: %s", subject)
	}
}

// startMongoDBBuildStatusSubscriber starts listening to NATS build statuses and saving them to MongoDB
func startMongoDBBuildStatusSubscriber(natsClient *common.NATSClient, mongoService *services.MongoLogService) {
	subject := "mira.status.*"
	_, err := natsClient.GetConnection().Subscribe(subject, func(msg *nats.Msg) {
		var buildStatus common.BuildStatus
		if err := json.Unmarshal(msg.Data, &buildStatus); err != nil {
			log.Printf("Failed to unmarshal build status: %v", err)
			return
		}

		// Save build status with project_id and app_name from the build status
		err := mongoService.SaveBuildStatus(
			buildStatus.BuildID,
			buildStatus.ProjectID,
			buildStatus.AppName,
			buildStatus.Status,
			buildStatus.StartedAt,
			buildStatus.CompletedAt,
			buildStatus.Error,
			buildStatus.ImageName,
		)
		if err != nil {
			log.Printf("Failed to save build status to MongoDB: %v", err)
		} else {
			log.Printf("✅ Saved build status to MongoDB for build %s: %s (Project: %s, App: %s)",
				buildStatus.BuildID, buildStatus.Status, buildStatus.ProjectID, buildStatus.AppName)
		}
	})
	if err != nil {
		log.Printf("Failed to subscribe to build statuses: %v", err)
	} else {
		log.Printf("Started MongoDB build status subscriber on subject: %s", subject)
	}
}
