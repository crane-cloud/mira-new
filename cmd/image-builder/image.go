package imagebuilder

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/buildpacks/pack/pkg/client"
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

	captureAndStreamLogs(
		func() {
			err := BuildImage(app)
			if err != nil {
				log.Fatalf("failed to build image: %v", err)
			}
			fmt.Println("Image built successfully")
		},
		func(line string) {
			logger.Log(map[string]string{
				"event": "buildpack",
			}, line)
		})

	return nil
}

func BuildImage(app *cTypes.Application) error {
	cli, err := client.NewClient()
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

func captureAndStreamLogs(f func(), streamFn func(string)) {
	origStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Start a goroutine to read and stream each line
	done := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(r)
		for scanner.Scan() {
			line := scanner.Text()
			streamFn(line)
		}
		close(done)
	}()

	f()

	// Close writer and restore stdout
	w.Close()
	os.Stdout = origStdout
	<-done
}
