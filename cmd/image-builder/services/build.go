package services

import (
	"context"
	"log"

	common "mira/cmd/common"
	"mira/cmd/image-builder/models"
	imageUtils "mira/cmd/image-builder/utils"

	buildpackClient "github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/image"
	"github.com/buildpacks/pack/pkg/logging"
)

// BuildService handles image building operations
type BuildService struct{}

// NewBuildService creates a new build service
func NewBuildService() *BuildService {
	return &BuildService{}
}

// BuildImage builds a Docker image using buildpacks
func (b *BuildService) BuildImage(buildSpec *models.BuildSpec, sourcePath string, natsLogger common.Logger) error {
	natsLogger.InfoWithStep("build", "Image Build Process Started")

	logger := logging.NewLogWithWriters(natsLogger, natsLogger)

	cliClient, err := buildpackClient.NewClient(
		buildpackClient.WithLogger(logger),
	)
	if err != nil {
		log.Printf("failed to create pack client: %v", err)
		return err
	}

	// Prepare build configuration
	buildOpts, err := b.prepareBuildOptions(buildSpec, sourcePath)
	if err != nil {
		return err
	}

	// Execute build
	if err := cliClient.Build(context.Background(), buildOpts); err != nil {
		log.Printf("failed to build image: %v", err)
		return err
	}

	natsLogger.InfoWithStep("build", "SUCCESS: Image built successfully: "+buildSpec.Name)
	return nil
}

// prepareBuildOptions creates build options for the pack client
func (b *BuildService) prepareBuildOptions(buildSpec *models.BuildSpec, sourcePath string) (buildpackClient.BuildOptions, error) {
	imageName := imageUtils.GenerateImageName(buildSpec)

	// Base environment variables
	env := map[string]string{
		// Remove root user settings to avoid permission conflicts
		// "CNB_USER_ID":  "0", // Root user ID
		// "CNB_GROUP_ID": "0", // Root group ID
	}

	// Add custom environment variables
	if buildSpec.Spec.Env != nil {
		for key, value := range buildSpec.Spec.Env {
			env[key] = value
		}
	}

	// Determine buildpacks and builder based on SSR configuration
	var buildpacks []string
	var builder string

	if !buildSpec.Spec.SSR {
		// Static site configuration
		env["BP_WEB_SERVER"] = "httpd"
		// env["BP_WEB_SERVER"] = "nginx"
		env["BP_WEB_SERVER_FORCE_HTTPS_REDIRECT"] = "false"
		env["BP_NODE_RUN_SCRIPTS"] = buildSpec.Spec.BuildCommand
		env["BP_WEB_SERVER_ROOT"] = buildSpec.Spec.OutputDir
		// Enable SPA support for React apps
		env["BP_WEB_SERVER_ENABLE_PUSH_STATE"] = "true"
		// Use custom nginx configuration
		// env["BP_NGINX_CONF_LOCATION"] = "scripts/nginx.conf"
		env["NODE_ENV"] = "production"
		buildpacks = append(buildpacks, "paketo-buildpacks/web-servers")
		builder = "paketobuildpacks/builder-jammy-base"
	} else {
		// Server-side rendering configuration
		buildpacks = append(buildpacks, "heroku/nodejs")
		builder = "heroku/builder:24"
	}

	return buildpackClient.BuildOptions{
		AppPath:    sourcePath,
		Builder:    builder,
		Image:      imageName,
		PullPolicy: image.PullIfNotPresent,
		Publish:    true,
		Env:        env,
		Buildpacks: buildpacks,
	}, nil
}
