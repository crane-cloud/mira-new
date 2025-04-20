package imagebuilder

import (
	"context"
	"fmt"
	"log"

	"github.com/buildpacks/pack/pkg/client"
	"github.com/buildpacks/pack/pkg/logging"
	dLogger "github.com/open-ug/conveyor/pkg/driver-runtime/log"
	cTypes "github.com/open-ug/conveyor/pkg/types"
)

// CreateBuildpacksImage creates a buildpacks image
func CreateBuildpacksImage(app *cTypes.Application, logger *dLogger.DriverLogger) error {

	if app.Spec.Source.Type == "git" {
		err := CloneGitRepo(app)
		if err != nil {
			log.Fatalf("failed to clone git repository: %v", err)
			return err
		}
	} else if app.Spec.Source.Type == "file" {
		err := HandleFileSource(app)
		if err != nil {
			log.Fatalf("failed to download file: %v", err)
			return err
		}
	}

	fmt.Println("building image")

	err := BuildImage(app, logger)
	if err != nil {
		log.Fatalf("failed to build image: %v", err)
		return err
	}
	fmt.Println("Image built successfully")

	return nil
}

func BuildImage(app *cTypes.Application, driverLogger *dLogger.DriverLogger) error {

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
		AppPath: appPath,
		Builder: "heroku/builder:24",
		Image:   app.Name + "-bpimage",
		//PullPolicy: config.PullAlways,
	}

	if err := cli.Build(context.Background(), buildOpts); err != nil {
		log.Fatalf("failed to build image: %v", err)
		return err
	}

	return nil
}
