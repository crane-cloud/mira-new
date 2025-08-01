package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	common "mira/cmd/common"
	"mira/cmd/image-builder/models"
	imageUtils "mira/cmd/image-builder/utils"

	"github.com/go-resty/resty/v2"
)

// DeployService handles deployment operations
type DeployService struct{}

// NewDeployService creates a new deployment service
func NewDeployService() *DeployService {
	return &DeployService{}
}

// DeployToCraneCloud deploys the built image to Crane Cloud
func (d *DeployService) DeployToCraneCloud(buildSpec *models.BuildSpec, logger *common.NATSLogger) error {
	logger.InfoWithStep("deploy", "Deploying image to Crane Cloud: "+buildSpec.Name)

	deployConfig := imageUtils.CreateDeploymentConfig(buildSpec)

	err := d.postToCraneCloud(deployConfig, buildSpec.Spec.AccessToken)
	if err != nil {
		logger.ErrorWithStep("deploy", "Error deploying image to Crane Cloud")
		return fmt.Errorf("error deploying image to Crane Cloud: %w", err)
	}

	return nil
}

// postToCraneCloud sends the deployment request to Crane Cloud API
func (d *DeployService) postToCraneCloud(deployConfig *models.DeploymentConfig, accessToken string) error {
	ctx := context.Background()

	// Get Crane Cloud API host from environment
	ccApiHost := os.Getenv("CRANECLOUD_API_HOST")
	if ccApiHost == "" {
		return fmt.Errorf("CRANECLOUD_API_HOST environment variable not set")
	}

	// Prepare HTTP client
	client := resty.New()
	client.SetHeader("Content-Type", "application/json")

	// Marshal deployment configuration
	jsonMessage, err := json.Marshal(deployConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment config: %w", err)
	}

	// Make deployment request
	resp, err := client.R().
		SetBody(jsonMessage).
		SetContext(ctx).
		SetAuthToken(accessToken).
		Post(ccApiHost + "/projects/" + deployConfig.ProjectID + "/apps")

	if err != nil {
		return fmt.Errorf("failed to make deployment request: %w", err)
	}

	if resp.IsError() {
		return fmt.Errorf("deployment request failed with status %s: %s", resp.Status(), resp.String())
	}

	return nil
}
