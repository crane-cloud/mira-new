package imagebuilder

import (
	"context"
	"log"
	"os"

	common "mira/cmd/common"

	"github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/image"
	"github.com/buildpacks/pack/pkg/logging"
)

// CreateBuildpacksImage creates a buildpacks image
func CreateBuildpacksImage(app *ImageBuild, logger *common.NATSLogger) error {

	if app.Spec.Source.Type == "git" {
		logger.InfoWithStep("clone", "Fetching Codebase from Git Repository")
		err := CloneGitRepo(app)
		if err != nil {
			log.Printf("failed to clone git repository: %v", err)
			logger.ErrorWithStep("clone", "Failed to clone git repository")
			return err
		}
	} else if app.Spec.Source.Type == "file" {
		logger.InfoWithStep("download", "Downloading File from URL")
		err := HandleFileSource(app)
		if err != nil {
			log.Printf("failed to download file: %v", err)
			logger.ErrorWithStep("download", "Failed to download file")
			return err
		}
	}

	logger.InfoWithStep("build", "Image Build Process Started")

	err := BuildImage(app, logger)
	if err != nil {
		log.Printf("failed to build image: %v", err)
		logger.ErrorWithStep("build", "Image build failed")
		return err
	}
	logger.InfoWithStep("build", "SUCCESS: Image built successfully: "+app.Name)

	return nil
}

func BuildImage(app *ImageBuild, natsLogger *common.NATSLogger) error {

	logger := logging.NewLogWithWriters(natsLogger, natsLogger)

	cliClient, err := client.NewClient(
		client.WithLogger(logger),
	)
	if err != nil {
		log.Printf("failed to create pack client: %v", err)
		return err
	}

	var appPath string
	if app.Spec.Source.Type == "git" {
		appPath = "/usr/local/crane/git/" + app.Name
	} else if app.Spec.Source.Type == "file" {
		appPath = "/usr/local/crane/zip/" + app.Name
	}

	//WriteNginxConfig(appPath, app.Spec.OutputDir)

	var DOCKER_USERNAME = os.Getenv("DOCKERHUB_USERNAME")

	var env = map[string]string{
		"CNB_USER_ID":  "0", // Root user ID
		"CNB_GROUP_ID": "0", // Root group ID
	}

	var buildpacks = []string{}
	var builder string

	/*
		add these envs to already existing envs
		map[string]string{
				"BP_NODE_RUN_SCRIPTS": app.Spec.BuildCommand,
				"BP_WEB_SERVER_ROOT":  app.Spec.OutputDir,
				"BP_WEB_SERVER":       "httpd",
				"CNB_USER_ID":         "0", // Root user ID
				"CNB_GROUP_ID":        "0", // Root group ID
			},
	*/

	if app.Spec.Env != nil {
		for key, value := range app.Spec.Env {
			env[key] = value
		}
	}

	if !app.Spec.SSR {
		env["BP_WEB_SERVER"] = "httpd"
		env["BP_NODE_RUN_SCRIPTS"] = app.Spec.BuildCommand
		env["BP_WEB_SERVER_ROOT"] = app.Spec.OutputDir
		buildpacks = append(buildpacks, "paketo-buildpacks/web-servers")
		builder = "paketobuildpacks/builder-jammy-base"
	} else {
		buildpacks = append(buildpacks, "heroku/nodejs")
		builder = "heroku/builder:24"
	}

	// Build configuration: We are to use Paketo Buildpacks
	buildOpts := client.BuildOptions{
		AppPath:    appPath,
		Builder:    builder,
		Image:      DOCKER_USERNAME + "/" + app.Spec.ProjectID + app.Name,
		PullPolicy: image.PullIfNotPresent,
		Publish:    true,
		Env:        env,
		Buildpacks: buildpacks,
	}

	if err := cliClient.Build(context.Background(), buildOpts); err != nil {
		log.Printf("failed to build image: %v", err)
		return err
	}

	return nil
}
