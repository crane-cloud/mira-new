package schemas

import (
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// Validation constants
const (
	MaxNameLength         = 63 // Kubernetes name limit
	MaxBuildCommandLength = 500
	MaxOutputDirLength    = 255
	MaxProjectIDLength    = 100
	MaxTokenLength        = 500
	MaxEnvVarCount        = 50
	MaxEnvVarKeyLength    = 100
	MaxEnvVarValueLength  = 1000
)

// Validation patterns
var (
	// Kubernetes-compatible name pattern (DNS subdomain)
	validNamePattern = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)
	// Project ID pattern (alphanumeric with dashes/underscores)
	validProjectIdPattern = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[1-5][a-fA-F0-9]{3}-[89abAB][a-fA-F0-9]{3}-[a-fA-F0-9]{12}$`)
	// Safe build command pattern (no shell injection characters)
	dangerousCommandPattern = regexp.MustCompile(`[;&|<>$\x60\\]`)
)

// ValidationError represents a validation error with field context
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// GenerateImageRequest represents the JSON request body for containerization
type GenerateImageRequest struct {
	Name            string            `json:"name" example:"my-app" validate:"required" doc:"Application name"`
	BuildCommand    string            `json:"build_command" example:"npm run build" validate:"required" doc:"Build command to execute"`
	OutputDirectory string            `json:"output_directory" example:"dist" validate:"required" doc:"Output directory after build"`
	AccessToken     string            `json:"access_token" validate:"required" doc:"Crane Cloud authentication token"`
	ProjectId       string            `json:"project_id" example:"proj-123" validate:"required" doc:"Crane Cloud project ID"`
	SSR             bool              `json:"ssr" example:"false" doc:"Enable server-side rendering"`
	Env             map[string]string `json:"env" doc:"Environment variables for the build"`
	Repo            string            `json:"repo" example:"https://github.com/user/repo.git" validate:"required" doc:"Git repository URL"`
}

// Validation functions
func validateName(name string) error {
	if name == "" {
		return ValidationError{Field: "name", Message: "is required"}
	}
	if len(name) > MaxNameLength {
		return ValidationError{Field: "name", Message: fmt.Sprintf("must be %d characters or less", MaxNameLength)}
	}
	if !validNamePattern.MatchString(name) {
		return ValidationError{Field: "name", Message: "must contain only lowercase letters, numbers, and hyphens, and start/end with alphanumeric characters"}
	}
	return nil
}

func validateBuildCommand(buildCommand string) error {
	if buildCommand == "" {
		return ValidationError{Field: "build_command", Message: "is required"}
	}
	if len(buildCommand) > MaxBuildCommandLength {
		return ValidationError{Field: "build_command", Message: fmt.Sprintf("must be %d characters or less", MaxBuildCommandLength)}
	}
	if dangerousCommandPattern.MatchString(buildCommand) {
		return ValidationError{Field: "build_command", Message: "contains potentially dangerous shell characters"}
	}
	return nil
}

func validateOutputDirectory(outputDir string) error {
	if outputDir == "" {
		return ValidationError{Field: "output_directory", Message: "is required"}
	}
	if len(outputDir) > MaxOutputDirLength {
		return ValidationError{Field: "output_directory", Message: fmt.Sprintf("must be %d characters or less", MaxOutputDirLength)}
	}
	// Normalize and validate path
	cleanPath := filepath.Clean(outputDir)
	if strings.Contains(cleanPath, "..") {
		return ValidationError{Field: "output_directory", Message: "cannot contain parent directory references (..)"}
	}
	if filepath.IsAbs(cleanPath) {
		return ValidationError{Field: "output_directory", Message: "must be a relative path"}
	}
	return nil
}

func validateAccessToken(token string) error {
	if token == "" {
		return ValidationError{Field: "access_token", Message: "is required"}
	}
	if len(token) > MaxTokenLength {
		return ValidationError{Field: "access_token", Message: fmt.Sprintf("must be %d characters or less", MaxTokenLength)}
	}
	// Basic token format validation (adjust based on your token format)
	if len(strings.TrimSpace(token)) < 10 {
		return ValidationError{Field: "access_token", Message: "appears to be invalid (too short)"}
	}
	return nil
}

func validateProjectId(projectId string) error {
	if projectId == "" {
		return ValidationError{Field: "project_id", Message: "is required"}
	}
	if len(projectId) > MaxProjectIDLength {
		return ValidationError{Field: "project_id", Message: fmt.Sprintf("must be %d characters or less", MaxProjectIDLength)}
	}
	if !validProjectIdPattern.MatchString(projectId) {
		return ValidationError{Field: "project_id", Message: "must be a valid UUID"}
	}
	return nil
}

func validateGitRepo(repo string) error {
	if repo == "" {
		return ValidationError{Field: "repo", Message: "is required"}
	}

	// Parse URL
	parsedURL, err := url.Parse(repo)
	if err != nil {
		return ValidationError{Field: "repo", Message: "must be a valid URL"}
	}

	// Validate scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ValidationError{Field: "repo", Message: "must use http or https protocol"}
	}

	// Validate host
	if parsedURL.Host == "" {
		return ValidationError{Field: "repo", Message: "must have a valid host"}
	}

	// Basic validation for common git hosting services
	validHosts := []string{"github.com", "gitlab.com", "bitbucket.org"}
	isValidHost := false
	for _, validHost := range validHosts {
		if strings.Contains(parsedURL.Host, validHost) {
			isValidHost = true
			break
		}
	}

	if !isValidHost {
		return ValidationError{Field: "repo", Message: "must be from a supported git hosting service (GitHub, GitLab, or Bitbucket)"}
	}

	return nil
}

func validateEnvVars(env map[string]string) error {
	if len(env) > MaxEnvVarCount {
		return ValidationError{Field: "env", Message: fmt.Sprintf("cannot have more than %d environment variables", MaxEnvVarCount)}
	}

	for key, value := range env {
		if key == "" {
			return ValidationError{Field: "env", Message: "environment variable keys cannot be empty"}
		}
		if len(key) > MaxEnvVarKeyLength {
			return ValidationError{Field: "env", Message: fmt.Sprintf("environment variable key '%s' is too long (max %d characters)", key, MaxEnvVarKeyLength)}
		}
		if len(value) > MaxEnvVarValueLength {
			return ValidationError{Field: "env", Message: fmt.Sprintf("environment variable value for '%s' is too long (max %d characters)", key, MaxEnvVarValueLength)}
		}

		// Validate key format (should be valid env var name)
		if !regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`).MatchString(key) {
			return ValidationError{Field: "env", Message: fmt.Sprintf("environment variable key '%s' contains invalid characters", key)}
		}
	}

	return nil
}

// ValidateGenerateImageRequest performs comprehensive validation
func ValidateGenerateImageRequest(req *GenerateImageRequest) []ValidationError {
	var errors []ValidationError

	// Validate each field
	if err := validateName(req.Name); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	if err := validateBuildCommand(req.BuildCommand); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	if err := validateOutputDirectory(req.OutputDirectory); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	if err := validateAccessToken(req.AccessToken); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	if err := validateProjectId(req.ProjectId); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Validate git repository
	if err := validateGitRepo(req.Repo); err != nil {
		errors = append(errors, err.(ValidationError))
	}

	// Validate environment variables
	if req.Env != nil {
		if err := validateEnvVars(req.Env); err != nil {
			errors = append(errors, err.(ValidationError))
		}
	}

	return errors
}
