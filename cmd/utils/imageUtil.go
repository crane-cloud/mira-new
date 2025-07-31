package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	internalsUtils "mira/internals/utils"
)

// ImageConfig represents configuration for image operations
type ImageConfig struct {
	Name         string            `json:"name"`
	ProjectID    string            `json:"project_id"`
	BuildCommand string            `json:"build_command"`
	OutputDir    string            `json:"output_dir"`
	Environment  map[string]string `json:"environment"`
}

// ValidateImageConfig validates image configuration
func ValidateImageConfig(config *ImageConfig) error {
	if config.Name == "" {
		return fmt.Errorf("image name is required")
	}
	if config.ProjectID == "" {
		return fmt.Errorf("project ID is required")
	}
	return nil
}

// GenerateDockerImageName creates a standardized Docker image name
func GenerateDockerImageName(projectID, imageName string) string {
	dockerUsername := os.Getenv("DOCKERHUB_USERNAME")
	if dockerUsername == "" {
		dockerUsername = "default"
	}
	return fmt.Sprintf("%s/%s%s", dockerUsername, projectID, imageName)
}

// CleanupBuildArtifacts removes temporary build files
func CleanupBuildArtifacts(paths ...string) error {
	for _, path := range paths {
		if path != "" {
			if err := os.RemoveAll(path); err != nil {
				return fmt.Errorf("failed to cleanup %s: %w", path, err)
			}
		}
	}
	return nil
}

// CreateBuildDirectory creates a build directory with proper permissions
func CreateBuildDirectory(basePath, name string) (string, error) {
	buildDir := filepath.Join(basePath, name)

	// Remove existing directory if it exists
	if _, err := os.Stat(buildDir); err == nil {
		if err := os.RemoveAll(buildDir); err != nil {
			return "", fmt.Errorf("failed to remove existing directory: %w", err)
		}
	}

	// Create directory with proper permissions
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create build directory: %w", err)
	}

	return buildDir, nil
}

// ParseEnvironmentVariables parses environment variables from JSON string
func ParseEnvironmentVariables(envJSON string) (map[string]string, error) {
	if envJSON == "" {
		return make(map[string]string), nil
	}

	return internalsUtils.ParseJSONToMap(envJSON)
}

// SanitizeImageName ensures image name follows Docker naming conventions
func SanitizeImageName(name string) string {
	// Convert to lowercase
	name = strings.ToLower(name)

	// Replace invalid characters with hyphens
	var result strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.' {
			result.WriteRune(r)
		} else {
			result.WriteRune('-')
		}
	}

	// Remove leading/trailing hyphens and ensure it doesn't start with a number
	sanitized := strings.Trim(result.String(), "-")
	if len(sanitized) > 0 && sanitized[0] >= '0' && sanitized[0] <= '9' {
		sanitized = "img-" + sanitized
	}

	return sanitized
}
