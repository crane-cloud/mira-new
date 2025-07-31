package handlers

import (
	"strings"
	"time"

	"fmt"
	_ "mira/cmd/api/models"
	common "mira/cmd/common"
	utils "mira/internals/utils"

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

// GenerateImageRequest represents the JSON request body for containerization
type GenerateImageRequest struct {
	Name            string            `json:"name" example:"my-app" validate:"required" doc:"Application name"`
	Type            string            `json:"type" example:"git" validate:"required" doc:"Source type: 'git' or 'file'"`
	BuildCommand    string            `json:"build_command" example:"npm run build" validate:"required" doc:"Build command to execute"`
	OutputDirectory string            `json:"output_directory" example:"dist" validate:"required" doc:"Output directory after build"`
	Token           string            `json:"token" validate:"required" doc:"Crane Cloud authentication token"`
	Project         string            `json:"project" example:"proj-123" validate:"required" doc:"Crane Cloud project ID"`
	SSR             bool              `json:"ssr" example:"false" doc:"Enable server-side rendering"`
	Env             map[string]string `json:"env" doc:"Environment variables for the build"`
	Repo            string            `json:"repo,omitempty" example:"https://github.com/user/repo.git" doc:"Git repository URL (required for git type)"`
	FileURL         string            `json:"file_url,omitempty" example:"http://example.com/file.zip" doc:"File URL (required for file type when using JSON)"`
}

// GenerateImage containerizes source code into Docker images
// @Summary Containerize source code
// @Description Converts source code from Git repository or uploaded file into a Docker image and deploys to Crane Cloud
// @Tags images
// @Accept json,multipart/form-data
// @Produce json
// @Param request body GenerateImageRequest true "Build configuration"
// @Success 200 {object} models.BuildResponse "Build started successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid request"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /images/containerize [post]
func (h *ImageHandler) GenerateImage(c *fiber.Ctx) error {
	contentType := string(c.Request().Header.ContentType())

	// Generate unique build ID
	buildID := uuid.New().String()

	var buildReq common.BuildRequest
	buildReq.ID = buildID
	buildReq.Timestamp = time.Now()

	// Handle JSON requests
	if strings.Contains(contentType, "application/json") {
		var req GenerateImageRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON format",
			})
		}

		// Validate required fields
		if req.Name == "" || req.Type == "" || req.BuildCommand == "" || req.OutputDirectory == "" || req.Token == "" || req.Project == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required fields: name, type, build_command, output_directory, token, project",
			})
		}

		if req.Type != "git" && req.Type != "file" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid source type. Must be 'git' or 'file'",
			})
		}

		// Map JSON fields to build request structure
		buildReq.Name = req.Name
		buildReq.Spec.BuildCommand = req.BuildCommand
		buildReq.Spec.OutputDir = req.OutputDirectory
		buildReq.Spec.ProjectID = req.Project
		buildReq.Spec.Token = req.Token
		buildReq.Spec.SSR = req.SSR
		buildReq.Spec.Env = req.Env
		if buildReq.Spec.Env == nil {
			buildReq.Spec.Env = make(map[string]string)
		}

		// Handle source type
		if req.Type == "git" {
			if req.Repo == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "repo field is required for git source type",
				})
			}
			buildReq.Spec.Source.GitRepo.URL = req.Repo
			buildReq.Spec.Source.Type = "git"
		} else if req.Type == "file" {
			if req.FileURL == "" {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "file_url field is required for file source type when using JSON",
				})
			}
			buildReq.Spec.Source.BlobFile.Source = req.FileURL
			buildReq.Spec.Source.Type = "file"
		}

	} else if strings.Contains(contentType, "multipart/form-data") {
		// Handle multipart/form-data requests (existing logic)
		form, err := c.MultipartForm()
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid form-data format",
			})
		}

		// get the fields from the form
		name := form.Value["name"]
		sourceType := form.Value["type"]
		buildCmd := form.Value["build_command"]
		outputDir := form.Value["output_directory"]
		token := form.Value["token"]
		project := form.Value["project"]

		if len(name) == 0 || len(sourceType) == 0 || len(buildCmd) == 0 || len(outputDir) == 0 || len(token) == 0 || len(project) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Missing required fields: name, type, build_command, output_directory, token, project",
			})
		}

		if sourceType[0] != "git" && sourceType[0] != "file" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid source type. Must be 'git' or 'file'",
			})
		}

		buildReq.Name = name[0]
		buildReq.Spec.BuildCommand = buildCmd[0]
		buildReq.Spec.OutputDir = outputDir[0]
		buildReq.Spec.ProjectID = project[0]
		buildReq.Spec.Token = token[0]
		buildReq.Spec.SSR = len(form.Value["ssr"]) > 0 && form.Value["ssr"][0] == "true"
		buildReq.Spec.Env = make(map[string]string)

		// get the environment variables from the form
		envVars := form.Value["env"]
		if len(envVars) > 0 {
			envMap, err := utils.ParseJSONToMap(envVars[0])
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid environment variables format",
				})
			}
			buildReq.Spec.Env = envMap
		}

		// Handle file upload for form-data
		if sourceType[0] == "file" {
			file, err := c.FormFile("file")
			if err != nil {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid file format",
				})
			}

			// Generate a unique name for the file
			uniqueName := fmt.Sprintf("%s-%s-%s", utils.GenerateRandomString(6), buildReq.Name, file.Filename)

			// save the file to the server
			err = c.SaveFile(file, "./uploads/"+uniqueName)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to save file",
				})
			}

			// get server url protocol and host
			serverURL := fmt.Sprintf("%s://%s", c.Protocol(), string(c.Context().URI().Host()))
			buildReq.Spec.Source.BlobFile.Source = serverURL + "/uploads/" + uniqueName
			buildReq.Spec.Source.Type = "file"
		} else if sourceType[0] == "git" {
			// get the git fields from the form
			repo := form.Value["repo"]
			if len(repo) == 0 {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "repo field is required for git source type",
				})
			}
			buildReq.Spec.Source.GitRepo.URL = repo[0]
			buildReq.Spec.Source.Type = "git"
		}
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Content-Type must be either application/json or multipart/form-data",
		})
	}

	// Publish build request to NATS
	err := h.natsClient.PublishBuildRequest(&buildReq)
	if err != nil {
		fmt.Printf("Failed to publish build request: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to queue build request",
		})
	}

	// Get host for WebSocket URL
	host := string(c.Context().URI().Host())
	if host == "" {
		host = "localhost:3000"
	}

	return c.JSON(fiber.Map{
		"message": "Image generation started",
		"data": fiber.Map{
			"name":     buildReq.Name,
			"build_id": buildReq.ID,
			"logs_url": getWebSocketURL(host, buildReq.ID),
		},
	})

}
