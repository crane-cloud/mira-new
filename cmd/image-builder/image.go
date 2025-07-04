package imagebuilder

import (
	"context"
	"fmt"
	"log"

	"github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/image"
	"github.com/buildpacks/pack/pkg/logging"
	dLogger "github.com/open-ug/conveyor/pkg/driver-runtime/log"
)

// CreateBuildpacksImage creates a buildpacks image
func CreateBuildpacksImage(app *ImageBuild, logger *dLogger.DriverLogger) error {

	if app.Spec.Source.Type == "git" {
		logger.Log(map[string]string{}, "Fetching Codebase from Git Repository")
		err := CloneGitRepo(app)
		if err != nil {
			log.Fatalf("failed to clone git repository: %v", err)
			return err
		}
	} else if app.Spec.Source.Type == "file" {
		logger.Log(map[string]string{}, "Downloading File from URL")
		err := HandleFileSource(app)
		if err != nil {
			log.Fatalf("failed to download file: %v", err)
			return err
		}
	}

	fmt.Println("building image")
	logger.Log(map[string]string{}, "Image Build Process Started")

	err := BuildImage(app, logger)
	if err != nil {
		log.Fatalf("failed to build image: %v", err)
		return err
	}
	fmt.Println("Image built successfully")
	logger.Log(map[string]string{}, "SUCCESS: Image built successfully: "+app.Name)

	return nil
}

func BuildImage(app *ImageBuild, driverLogger *dLogger.DriverLogger) error {

	logger := logging.NewLogWithWriters(driverLogger, driverLogger)

	cli, err := client.NewClient(
		client.WithLogger(logger),
	)
	if err != nil {
		log.Fatalf("failed to create pack client: %v", err)
		return err
	}

	var appPath string
	if app.Spec.Source.Type == "git" {
		appPath = "/usr/local/crane/git/" + app.Name
	} else if app.Spec.Source.Type == "file" {
		appPath = "/usr/local/crane/zip/" + app.Name
	}

	buildOpts := client.BuildOptions{
		AppPath:    appPath,
		Builder:    "heroku/builder:24",
		Image:      app.Name + "-bpimage",
		PullPolicy: image.PullIfNotPresent,
	}

	if err := cli.Build(context.Background(), buildOpts); err != nil {
		log.Fatalf("failed to build image: %v", err)
		return err
	}

	return nil
}
