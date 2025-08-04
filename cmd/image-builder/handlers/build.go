package handlers

import (
	"fmt"
	"log"
	"time"

	common "mira/cmd/common"
	"mira/cmd/image-builder/models"
	"mira/cmd/image-builder/services"
	imageUtils "mira/cmd/image-builder/utils"
)

// BuildHandler handles build orchestration
type BuildHandler struct {
	gitService    *services.GitService
	buildService  *services.BuildService
	deployService *services.DeployService
	natsClient    *common.NATSClient
}

// NewBuildHandler creates a new build handler with all required services
func NewBuildHandler(natsClient *common.NATSClient) *BuildHandler {
	return &BuildHandler{
		gitService:    services.NewGitService(),
		buildService:  services.NewBuildService(),
		deployService: services.NewDeployService(),
		natsClient:    natsClient,
	}
}

// ProcessBuildRequest handles a build request received from NATS
func (h *BuildHandler) ProcessBuildRequest(buildReq *common.BuildRequest) error {
	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("MIRA Processing build request: %s", buildReq.ID)

	// Create NATS logger for this build
	logger := common.NewNATSLogger(h.natsClient.GetConnection(), buildReq.ID)

	// Publish build status: started
	status := &models.BuildStatus{
		BuildID:   buildReq.ID,
		Status:    "running",
		StartedAt: time.Now(),
	}
	h.natsClient.PublishBuildStatus((*common.BuildStatus)(status))

	// Convert BuildRequest to internal build spec
	buildSpec := imageUtils.ConvertToBuildSpec(buildReq)

	// Execute build pipeline
	err := h.executeBuildPipeline(buildSpec, logger)
	if err != nil {
		log.Printf("Error creating image: %v", err)
		logger.ErrorWithStep("build", fmt.Sprintf("Build failed: %v", err))

		// Update status: failed
		status.Status = "failed"
		status.CompletedAt = time.Now()
		status.Error = err.Error()
		h.natsClient.PublishBuildStatus((*common.BuildStatus)(status))

		return fmt.Errorf("error creating image: %v", err)
	}

	// Log successful completion
	imageName := imageUtils.GenerateImageName(buildSpec)
	logger.InfoWithStep("deploy", "SUCCESSFULLY DEPLOYED IMAGE TO CRANE CLOUD: "+imageName)
	log.Printf("Image created and deployed successfully: %s", buildSpec.Name)

	// Update status: completed
	status.Status = "completed"
	status.CompletedAt = time.Now()
	status.ImageName = imageName
	h.natsClient.PublishBuildStatus((*common.BuildStatus)(status))

	return nil
}

// executeBuildPipeline runs the complete build pipeline
func (h *BuildHandler) executeBuildPipeline(buildSpec *models.BuildSpec, logger *common.NATSLogger) error {
	// Step 1: Handle source code (git clone or file download)
	sourcePath, err := h.handleSourceCode(buildSpec, logger)
	if err != nil {
		return fmt.Errorf("source handling failed: %w", err)
	}

	// Step 2: Build the image
	err = h.buildService.BuildImage(buildSpec, sourcePath, logger)
	if err != nil {
		return fmt.Errorf("image build failed: %w", err)
	}

	// Step 3: Deploy to Crane Cloud
	err = h.deployService.DeployToCraneCloud(buildSpec, logger)
	if err != nil {
		return fmt.Errorf("deployment failed: %w", err)
	}

	return nil
}

// handleSourceCode handles git cloning or file downloading based on source type
func (h *BuildHandler) handleSourceCode(buildSpec *models.BuildSpec, logger *common.NATSLogger) (string, error) {
	switch buildSpec.Source.Type {
	case "git":
		logger.InfoWithStep("clone", "Fetching Codebase from Git Repository")
		return h.gitService.CloneRepository(buildSpec, logger)
	case "file":
		logger.InfoWithStep("download", "Downloading File from URL")
		return h.gitService.HandleFileSource(buildSpec, logger)
	default:
		return "", fmt.Errorf("unsupported source type: %s", buildSpec.Source.Type)
	}
}
