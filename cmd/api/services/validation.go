package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
)

// ValidationService handles validation operations for the API
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
func (v *ValidationService) ValidateAppName(appName, projectID, accessToken string) error {
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
		SetAuthToken(accessToken).
		SetQueryParam("name", appName).
		Get(ccApiHost + "/projects/" + projectID + "/apps")

	if err != nil {
		return fmt.Errorf("failed to validate app name: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("app name validation request failed with status %s: %s", resp.Status(), resp.String())
	}

	// Parse response
	var appResponse AppExistsResponse
	if err := json.Unmarshal(resp.Body(), &appResponse); err != nil {
		// Log the actual response for debugging
		fmt.Printf("Failed to parse validation response: %v\n", err)
		fmt.Printf("Response body: %s\n", string(resp.Body()))
		return fmt.Errorf("failed to parse validation response: %w", err)
	}

	// Check if app with the same name exists
	// Try to parse the data as a map to check for apps
	if dataMap, ok := appResponse.Data.(map[string]interface{}); ok {
		// Try different possible structures
		if apps, exists := dataMap["apps"]; exists {
			if appsArray, ok := apps.([]interface{}); ok {
				for _, app := range appsArray {
					if appMap, ok := app.(map[string]interface{}); ok {
						if name, exists := appMap["name"]; exists {
							if appNameStr, ok := name.(string); ok && appNameStr == appName {
								return fmt.Errorf("app with name '%s' already exists in project", appName)
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
						if appNameStr, ok := name.(string); ok && appNameStr == appName {
							return fmt.Errorf("app with name '%s' already exists in project", appName)
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
					if appNameStr, ok := name.(string); ok && appNameStr == appName {
						return fmt.Errorf("app with name '%s' already exists in project", appName)
					}
				}
			}
		}
	}

	return nil
}
