package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	common "mira/cmd/common"
	"mira/cmd/image-builder/models"

	"github.com/go-resty/resty/v2"
)

// ValidationService handles validation operations
type ValidationService struct{}

// NewValidationService creates a new validation service
func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// AppExistsResponse represents the response from CraneCloud API when checking if app exists
type AppExistsResponse struct {
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

// ValidateAppName checks if an app with the given name already exists in the project
func (v *ValidationService) ValidateAppName(buildSpec *models.BuildSpec, logger common.Logger) error {
	logger.InfoWithStep("validation", "Validating app name: "+buildSpec.Name)

	// Get Crane Cloud API host from environment
	ccApiHost := os.Getenv("CRANECLOUD_API_HOST")
	if ccApiHost == "" {
		return fmt.Errorf("CRANECLOUD_API_HOST environment variable not set")
	}

	// Prepare HTTP client
	client := resty.New()
	client.SetHeader("Content-Type", "application/json")

	// Make API call to check if app exists
	ctx := context.Background()
	resp, err := client.R().
		SetContext(ctx).
		SetAuthToken(buildSpec.Spec.AccessToken).
		SetQueryParam("name", buildSpec.Name).
		Get(ccApiHost + "/projects/" + buildSpec.Spec.ProjectID + "/apps")

	if err != nil {
		logger.ErrorWithStep("validation", "Failed to validate app name: "+err.Error())
		return fmt.Errorf("failed to validate app name: %w", err)
	}

	if resp.IsError() {
		logger.ErrorWithStep("validation", "App name validation request failed with status: "+resp.Status())
		return fmt.Errorf("app name validation request failed with status %s: %s", resp.Status(), resp.String())
	}

	// Parse response
	var appResponse AppExistsResponse
	if err := json.Unmarshal(resp.Body(), &appResponse); err != nil {
		logger.ErrorWithStep("validation", "Failed to parse validation response: "+err.Error())
		logger.ErrorWithStep("validation", "Response body: "+string(resp.Body()))
		return fmt.Errorf("failed to parse validation response: %w", err)
	}

	// Log the parsed response structure for debugging
	logger.InfoWithStep("validation", "Parsed response data type: "+fmt.Sprintf("%T", appResponse.Data))

	// Check if app with the same name exists
	// Try to parse the data as a map to check for apps
	if dataMap, ok := appResponse.Data.(map[string]interface{}); ok {
		// Try different possible structures
		if apps, exists := dataMap["apps"]; exists {
			if appsArray, ok := apps.([]interface{}); ok {
				for _, app := range appsArray {
					if appMap, ok := app.(map[string]interface{}); ok {
						if name, exists := appMap["name"]; exists {
							if appName, ok := name.(string); ok && appName == buildSpec.Name {
								logger.ErrorWithStep("validation", "App with name '"+buildSpec.Name+"' already exists in project")
								return fmt.Errorf("app with name '%s' already exists in project", buildSpec.Name)
							}
						}
					}
				}
			}
		}
		// Try direct array structure
		if appsArray, ok := dataMap["data"].([]interface{}); ok {
			for _, app := range appsArray {
				if appMap, ok := app.(map[string]interface{}); ok {
					if name, exists := appMap["name"]; exists {
						if appName, ok := name.(string); ok && appName == buildSpec.Name {
							logger.ErrorWithStep("validation", "App with name '"+buildSpec.Name+"' already exists in project")
							return fmt.Errorf("app with name '%s' already exists in project", buildSpec.Name)
						}
					}
				}
			}
		}
	}
	// Try direct array structure
	if appsArray, ok := appResponse.Data.([]interface{}); ok {
		for _, app := range appsArray {
			if appMap, ok := app.(map[string]interface{}); ok {
				if name, exists := appMap["name"]; exists {
					if appName, ok := name.(string); ok && appName == buildSpec.Name {
						logger.ErrorWithStep("validation", "App with name '"+buildSpec.Name+"' already exists in project")
						return fmt.Errorf("app with name '%s' already exists in project", buildSpec.Name)
					}
				}
			}
		}
	}

	logger.InfoWithStep("validation", "App name validation passed: "+buildSpec.Name)
	return nil
}
