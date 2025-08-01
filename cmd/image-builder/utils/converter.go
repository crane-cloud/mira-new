package utils

import (
	"os"

	common "mira/cmd/common"
	"mira/cmd/image-builder/models"
)

// ConvertToBuildSpec converts a common.BuildRequest to internal BuildSpec
func ConvertToBuildSpec(buildReq *common.BuildRequest) *models.BuildSpec {
	return &models.BuildSpec{
		Name:   buildReq.Name,
		Spec:   buildReq.Spec,
		Source: buildReq.Spec.Source,
	}
}

// GenerateImageName creates a standardized image name
func GenerateImageName(buildSpec *models.BuildSpec) string {
	dockerUsername := os.Getenv("DOCKERHUB_USERNAME")
	return dockerUsername + "/" + buildSpec.Spec.ProjectID + buildSpec.Name
}

// CreateDeploymentConfig creates deployment configuration
func CreateDeploymentConfig(buildSpec *models.BuildSpec) *models.DeploymentConfig {
	imageName := GenerateImageName(buildSpec)

	envVars := map[string]string{
		"PORT": "8080",
	}

	// Add custom environment variables
	if buildSpec.Spec.Env != nil {
		for key, value := range buildSpec.Spec.Env {
			envVars[key] = value
		}
	}

	return &models.DeploymentConfig{
		Image:        imageName,
		Name:         buildSpec.Name,
		ProjectID:    buildSpec.Spec.ProjectID,
		PrivateImage: false,
		Replicas:     1,
		Port:         8080,
		EnvVars:      envVars,
	}
}
