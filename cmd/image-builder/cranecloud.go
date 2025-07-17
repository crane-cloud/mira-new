package imagebuilder

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
)

type CraneCloudData struct {
	EnvVars      map[string]string `json:"env_vars"`
	Image        string            `json:"image"`
	Name         string            `json:"name"`
	ProjectID    string            `json:"project_id"`
	PrivateImage bool              `json:"private_image"`
	Replicas     int               `json:"replicas"`
	Port         int               `json:"port"`
}

// Post to crane cloud
func (c *CraneCloudData) DeployToCraneCloud(token string) error {
	ctx := context.Background()

	// Crane cloud API HOST
	ccApiHost := os.Getenv("CRANECLOUD_API_HOST")

	client := resty.New()
	client.SetHeader("Content-Type", "application/json")
	jsonMessage, err := json.Marshal(c)
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}
	_, err = client.R().
		SetBody(jsonMessage).
		SetContext(ctx).
		SetAuthToken(token).
		Post(ccApiHost + "/projects/" + c.ProjectID + "/apps")
	if err != nil {
		fmt.Println("Error: ", err)
		return err
	}
	return nil
}
