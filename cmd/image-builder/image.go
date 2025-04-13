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
	// clone Source code
	fmt.Println("cloning repo")
	err := CloneGitRepo(app)

	if err != nil {
		log.Fatalf("failed to create pack client: %v", err)
		return err
	}
	fmt.Println("building image")

	captureAndStreamLogs(
		func() {
			err = BuildImage(app)
		},
		func(line string) {
			logger.Log(map[string]string{
				"event": "buildpack",
			}, line)
		})

	if err != nil {
		log.Fatalf("failed to create pack client: %v", err)
		return err
	}

	return nil
}

func BuildImage(app *cTypes.Application) error {
	cli, err := client.NewClient()
	if err != nil {
		log.Fatalf("failed to create pack client: %v", err)
		return err
	}

	buildOpts := client.BuildOptions{
		AppPath: "/usr/local/crane/git/" + app.Name,
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
