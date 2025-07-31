package imagebuilder

import (
	"fmt"
	"log"
	"os"
	"time"

	common "mira/cmd/common"
)

// ProcessBuildRequest handles a build request received from NATS
func ProcessBuildRequest(buildReq *common.BuildRequest, natsClient *common.NATSClient) error {
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("MIRA Processing build request: %s", buildReq.ID)

	// Create NATS logger for this build
	logger := common.NewNATSLogger(natsClient.GetConnection(), buildReq.ID)

	// Publish build status: started
	status := &common.BuildStatus{
		BuildID:   buildReq.ID,
		Status:    "running",
		StartedAt: time.Now(),
	}
	natsClient.PublishBuildStatus(status)

	// Convert BuildRequest to ImageBuild for compatibility
	imageSpec := ImageBuild{
		Name: buildReq.Name,
		Spec: ImageBuilderSpec{
			Source:       buildReq.Spec.Source,
			BuildCommand: buildReq.Spec.BuildCommand,
			OutputDir:    buildReq.Spec.OutputDir,
			ProjectID:    buildReq.Spec.ProjectID,
			Token:        buildReq.Spec.Token,
			SSR:          buildReq.Spec.SSR,
			Port:         buildReq.Spec.Port,
			Env:          buildReq.Spec.Env,
		},
	}

	err := CreateBuildpacksImage(&imageSpec, logger)
	if err != nil {
		log.Printf("Error creating image: %v", err)
		logger.ErrorWithStep("build", fmt.Sprintf("Build failed: %v", err))

		// Update status: failed
		status.Status = "failed"
		status.CompletedAt = time.Now()
		status.Error = err.Error()
		natsClient.PublishBuildStatus(status)

		return fmt.Errorf("error creating image: %v", err)
	}

	var DOCKER_USERNAME = os.Getenv("DOCKERHUB_USERNAME")
	imageName := DOCKER_USERNAME + "/" + imageSpec.Spec.ProjectID + imageSpec.Name

	// Deploy to Crane Cloud
	crane_cloud_app := &CraneCloudData{
		Image:        imageName,
		Name:         imageSpec.Name,
		ProjectID:    imageSpec.Spec.ProjectID,
		PrivateImage: false,
		Replicas:     1,
		Port:         8080,
		EnvVars: map[string]string{
			"PORT": "8080",
		},
	}

	logger.InfoWithStep("deploy", "Deploying image to Crane Cloud: "+imageSpec.Name)

	err = crane_cloud_app.DeployToCraneCloud(imageSpec.Spec.Token)
	if err != nil {
		logger.ErrorWithStep("deploy", "Error deploying image to Crane Cloud")

		// Update status: failed
		status.Status = "failed"
		status.CompletedAt = time.Now()
		status.Error = fmt.Sprintf("deployment failed: %v", err)
		natsClient.PublishBuildStatus(status)

		return fmt.Errorf("error deploying image to Crane Cloud: %v", err)
	}

	// Log successful completion
	logger.InfoWithStep("deploy", "SUCCESSFULLY DEPLOYED IMAGE TO CRANE CLOUD: "+imageName)
	log.Printf("Image created and deployed successfully: %s", imageSpec.Name)

	// Update status: completed
	status.Status = "completed"
	status.CompletedAt = time.Now()
	status.ImageName = imageName
	natsClient.PublishBuildStatus(status)

	return nil
}

func Listen() {
	// Create NATS client
	natsClient, err := common.NewNATSClient()
	if err != nil {
		fmt.Printf("Error creating NATS client: %v\n", err)
		return
	}
	defer natsClient.Close()

	fmt.Println("MIRA Image Builder started, listening for build requests...")

	// Subscribe to build requests
	_, err = natsClient.SubscribeToBuildRequests(func(buildReq *common.BuildRequest) {
		log.Printf("Received build request: %s", buildReq.ID)

		// Process build request in a goroutine to handle multiple requests concurrently
		go func() {
			err := ProcessBuildRequest(buildReq, natsClient)
			if err != nil {
				log.Printf("Build request %s failed: %v", buildReq.ID, err)
			} else {
				log.Printf("Build request %s completed successfully", buildReq.ID)
			}
		}()
	})

	if err != nil {
		fmt.Printf("Error subscribing to build requests: %v\n", err)
		return
	}

	// Keep the service running
	fmt.Println("Image builder is ready to process build requests")
	select {} // Block forever
}
