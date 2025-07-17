package imagebuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	c "github.com/open-ug/conveyor/pkg/client"
	runtime "github.com/open-ug/conveyor/pkg/driver-runtime"
	dLogger "github.com/open-ug/conveyor/pkg/driver-runtime/log"
)

// Listen for messages from the runtime
func Reconcile(payload string, event string, driverName string, logger *dLogger.DriverLogger) error {

	log.SetFlags(log.Ldate | log.Ltime)
	log.Printf("MIRA Reconciling: %v", payload)

	if event == "create" {
		var imageSpec ImageBuild
		err := json.Unmarshal([]byte(payload), &imageSpec)
		if err != nil {
			return fmt.Errorf("error unmarshalling application message: %v", err)
		}

		err = CreateBuildpacksImage(&imageSpec, logger)
		if err != nil {
			log.Printf("Error creating image from git repository: %v", err)
			return fmt.Errorf("error creating image from git repository: %v", err)
		}

		var DOCKER_USERNAME = os.Getenv("DOCKERHUB_USERNAME")

		// Deploy to Crane Cloud
		crane_cloud_app := &CraneCloudData{
			Image:        DOCKER_USERNAME + "/" + imageSpec.Spec.ProjectID + imageSpec.Name,
			Name:         imageSpec.Name,
			ProjectID:    imageSpec.Spec.ProjectID,
			PrivateImage: true,
			Replicas:     1,
			Port:         imageSpec.Spec.Port,
		}

		logger.Log(map[string]string{}, "Deploying image to Crane Cloud: "+imageSpec.Name)

		err = crane_cloud_app.DeployToCraneCloud(imageSpec.Spec.Token)
		if err != nil {
			logger.Log(map[string]string{}, "Error deploying image to Crane Cloud")
			return fmt.Errorf("error deploying image to Crane Cloud: %v", err)
		}
		// Log the successful creation of the image
		logger.Log(map[string]string{}, "SUCCESSFULLY DEPLOYED IMAGE TO CRANE CLOUD: "+crane_cloud_app.Image)

		log.Printf("Image created successfully: %s", imageSpec.Name)

	}

	return nil
}

func Listen() {
	client := c.NewClient()
	_, err := client.CreateOrUpdateResourceDefinition(context.Background(), ImageBuilder)

	if err != nil {
		fmt.Println("Error creating or updating resource definition: ", err)
	}

	driver := &runtime.Driver{
		Reconcile: Reconcile,
		Name:      "mira",
		Resources: []string{"image-builder"},
	}

	driverManager, err := runtime.NewDriverManager(driver, []string{"*"})
	if err != nil {
		fmt.Println("Error creating driver manager: ", err)
		return
	}

	err = driverManager.Run()
	if err != nil {
		fmt.Println("Error running driver manager: ", err)
	}

}
