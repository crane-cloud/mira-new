package handlers

import (
	"fmt"
	"time"

	_ "mira/cmd/api/models"
	"mira/cmd/api/schemas"
	common "mira/cmd/common"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type ImageHandler struct {
	natsClient *common.NATSClient
}

func NewImageHandler(natsClient *common.NATSClient) *ImageHandler {
	if natsClient == nil {
		var err error
		natsClient, err = common.NewNATSClient()
		if err != nil {
			fmt.Printf("Failed to create NATS client: %v\n", err)
			return nil
		}
	}
	return &ImageHandler{
		natsClient: natsClient,
	}
}

func getWebSocketURL(host, buildID string) string {
	return fmt.Sprintf("ws://%s/api/logs/%s", host, buildID)
}
func getLogsHTMLURL(host, buildID string) string {
	return fmt.Sprintf("http://%s/git-logs.html?buildId=%s", host, buildID)
}

// GenerateImage containerizes source code into Docker images
// @Summary Containerize source code
// @Description Converts source code from Git repository into a Docker image and deploys to Crane Cloud
// @Tags images
// @Accept json
// @Produce json
// @Param request body schemas.GenerateImageRequest true "Build configuration"
// @Success 200 {object} models.BuildResponse "Build started successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /images/containerize [post]
func (h *ImageHandler) GenerateImage(c *fiber.Ctx) error {
	// Generate unique build ID
	buildID := uuid.New().String()

	var buildReq common.BuildRequest
	buildReq.ID = buildID
	buildReq.Timestamp = time.Now()

	// Parse JSON request
	var req schemas.GenerateImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid JSON format",
			"details": err.Error(),
		})
	}

	// Comprehensive validation
	if validationErrors := schemas.ValidateGenerateImageRequest(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":      "Validation failed",
			"validation": validationErrors,
		})
	}

	// Map JSON fields to build request structure
	buildReq.Name = req.Name
	buildReq.Spec.BuildCommand = req.BuildCommand
	buildReq.Spec.OutputDir = req.OutputDirectory
	buildReq.Spec.ProjectID = req.ProjectId
	buildReq.Spec.AccessToken = req.AccessToken
	buildReq.Spec.SSR = req.SSR
	buildReq.Spec.Env = req.Env
	if buildReq.Spec.Env == nil {
		buildReq.Spec.Env = make(map[string]string)
	}

	// Set git repository as source
	buildReq.Spec.Source.GitRepo.URL = req.Repo
	buildReq.Spec.Source.Type = "git"

	// Get host for WebSocket URL first (before async operations)
	host := string(c.Context().URI().Host())
	if host == "" {
		host = "localhost:3000"
	}

	// Publish build request to NATS asynchronously for better response time
	published := make(chan error, 1)

	h.natsClient.PublishBuildRequestAsync(&buildReq,
		func() {
			// Success callback
			published <- nil
		},
		func(err error) {
			// Error callback
			fmt.Printf("Failed to publish build request asynchronously: %v\n", err)
			published <- err
		},
	)

	// Wait for publish result with timeout (non-blocking for client)
	select {
	case err := <-published:
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to queue build request",
				"details": err.Error(),
			})
		}
	case <-time.After(2 * time.Second):
		// Don't wait too long, respond optimistically
		fmt.Printf("Build request publish taking longer than expected for build %s\n", buildReq.ID)
	}

	return c.JSON(fiber.Map{
		"message": "Image generation started",
		"data": fiber.Map{
			"name":            buildReq.Name,
			"build_id":        buildReq.ID,
			"logs_socket_url": getWebSocketURL(host, buildReq.ID),
			"logs_html_url":   getLogsHTMLURL(host, buildReq.ID),
		},
	})
}
