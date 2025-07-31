package imagebuilder

import (
	"log"

	common "mira/cmd/common"
	"mira/cmd/image-builder/handlers"
)

// ProcessBuildRequest handles a build request received from NATS
// This function now delegates to the structured handler
func ProcessBuildRequest(buildReq *common.BuildRequest, natsClient *common.NATSClient) error {
	buildHandler := handlers.NewBuildHandler(natsClient)
	return buildHandler.ProcessBuildRequest(buildReq)
}

// Listen starts the image builder service and listens for build requests
func Listen() {
	// Create NATS client
	natsClient, err := common.NewNATSClient()
	if err != nil {
		log.Printf("Error creating NATS client: %v", err)
		return
	}
	defer natsClient.Close()

	log.Printf("MIRA Image Builder started, listening for build requests...")

	// Create build handler
	buildHandler := handlers.NewBuildHandler(natsClient)

	// Subscribe to build requests
	_, err = natsClient.SubscribeToBuildRequests(func(buildReq *common.BuildRequest) {
		log.Printf("Received build request: %s", buildReq.ID)

		// Process build request in a goroutine to handle multiple requests concurrently
		go func() {
			err := buildHandler.ProcessBuildRequest(buildReq)
			if err != nil {
				log.Printf("Build request %s failed: %v", buildReq.ID, err)
			} else {
				log.Printf("Build request %s completed successfully", buildReq.ID)
			}
		}()
	})

	if err != nil {
		log.Printf("Error subscribing to build requests: %v", err)
		return
	}

	// Keep the service running
	log.Printf("Image builder is ready to process build requests")
	select {} // Block forever
}
