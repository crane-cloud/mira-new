package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"mira/cmd/api/models"
	"mira/cmd/config"
	"mira/cmd/utils"

	"github.com/gofiber/fiber/v2"
)

// DetectFrameworkRequest represents the request body for framework detection
type DetectFrameworkRequest struct {
	RepoURL string `json:"repo_url" example:"https://github.com/user/repo" validate:"required"`
}

const (
	// Timeouts and limits
	requestTimeoutSeconds = 30

	// Rate limiting
	maxRequestsPerMinute = 60
)

// HTTP client with proper configuration
var httpClient = &http.Client{
	Timeout: requestTimeoutSeconds * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: false,
	},
}

// DetectFramework analyzes a GitHub repository to detect JavaScript frameworks
// @Summary Detect JavaScript framework from repository
// @Description Analyzes package.json and configuration files to detect JavaScript frameworks
// @Tags images
// @Accept json
// @Produce json
// @Param request body DetectFrameworkRequest true "Repository URL"
// @Success 200 {object} models.FrameworkDetectionResponse "Detected JavaScript frameworks"
// @Failure 400 {object} models.ErrorResponse "Invalid request or not a GitHub repository"
// @Failure 404 {object} models.ErrorResponse "Repository not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /images/detect [post]
func DetectFramework(c *fiber.Ctx) error {
	// Parse and validate request
	var req DetectFrameworkRequest
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: "Invalid request body format",
		})
	}

	// Validate and parse repository URL
	owner, repo, err := utils.ParseRepositoryURL(req.RepoURL)
	if err != nil {
		log.Printf("Error parsing repository URL %s: %v", req.RepoURL, err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error: fmt.Sprintf("Invalid repository URL: %v", err),
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeoutSeconds*time.Second)
	defer cancel()

	// First, check if repository exists and is accessible
	repoExists, err := checkRepositoryExists(ctx, owner, repo)
	if err != nil {
		log.Printf("Error checking repository %s/%s: %v", owner, repo, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to check repository",
		})
	}

	if !repoExists {
		log.Printf("Repository not found: %s/%s", owner, repo)
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error: "Repository not found. Please check the URL and ensure the repository exists.",
		})
	}

	// Clone repository to directory
	repoDir, cleanup, err := utils.CloneRepository(ctx, owner, repo)
	if err != nil {
		log.Printf("Error cloning repository %s/%s: %v", owner, repo, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to clone repository for analysis",
		})
	}
	// Ensure cleanup happens regardless of success or failure
	defer cleanup()

	// Check if it's a JavaScript project
	isJSProject, err := utils.IsJavaScriptProjectLocal(repoDir)
	if err != nil {
		log.Printf("Error checking if %s/%s is a JS project: %v", owner, repo, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to analyze repository",
		})
	}

	if !isJSProject {
		log.Printf("Repository %s/%s is not a JavaScript project", owner, repo)
		packageManager := utils.DetectPackageManager(repoDir)
		return c.Status(fiber.StatusOK).JSON(map[string]interface{}{
			"message":    "This repository does not appear to be a JavaScript application",
			"frameworks": map[string]interface{}{},
			"build_dir":  "build",
			"command":    packageManager + " build",
		})
	}

	// Detect JavaScript frameworks from local clone
	detectedFrameworks, err := utils.DetectJavaScriptFrameworksLocal(repoDir)
	if err != nil {
		log.Printf("Error detecting JS frameworks for %s/%s: %v", owner, repo, err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error: "Failed to analyze JavaScript frameworks",
		})
	}

	// Determine build settings based on detected frameworks and package manager
	primaryCommand, primaryBuildDir := utils.DetermineBuildInfo(detectedFrameworks, repoDir)

	// Prepare response
	frameworksMap := make(map[string]interface{})
	for _, framework := range detectedFrameworks {
		frameworkData := map[string]interface{}{
			"confidence":  framework.Confidence,
			"detected_in": framework.DetectedIn,
			"version":     framework.Version,
			"description": framework.Description,
		}

		frameworksMap[framework.Name] = frameworkData
	}

	var message string
	if len(detectedFrameworks) > 0 {
		message = "JavaScript frameworks detected successfully"
	} else {
		message = "This appears to be a JavaScript project, but no specific frameworks were detected"
	}

	response := map[string]interface{}{
		"message":    message,
		"frameworks": frameworksMap,
		"build_dir":  primaryBuildDir,
		"command":    primaryCommand,
	}

	log.Printf("Successfully detected %d JS frameworks for %s/%s", len(detectedFrameworks), owner, repo)
	return c.JSON(response)
}

// checkRepositoryExists verifies if the GitHub repository exists and is accessible
func checkRepositoryExists(ctx context.Context, owner, repo string) (bool, error) {
	apiURL := fmt.Sprintf("%s/repos/%s/%s", config.GITHUB_API_URL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "MIRA-Framework-Detection/1.0")

	// Add authentication if token is available
	if config.GITHUB_ACCESS_TOKEN != "" {
		req.Header.Set("Authorization", "Bearer "+config.GITHUB_ACCESS_TOKEN)
	}

	resp, err := httpClient.Do(req)

	if err != nil {
		return false, fmt.Errorf("failed to check repository: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNotFound:
		return false, nil
	case http.StatusForbidden:
		return false, fmt.Errorf("access forbidden - repository may be private")
	default:
		return false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}
