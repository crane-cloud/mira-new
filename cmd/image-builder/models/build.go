package models

import (
	"time"

	common "mira/cmd/common"
)

// BuildSpec represents the internal build specification
type BuildSpec struct {
	Name   string                    `json:"name"`
	Spec   common.ImageBuilderSpec   `json:"spec"`
	Source common.ImageBuilderSource `json:"source"`
}

// BuildStatus represents the status of a build operation
type BuildStatus struct {
	BuildID     string    `json:"build_id"`
	ProjectID   string    `json:"project_id,omitempty"`
	AppName     string    `json:"app_name,omitempty"`
	Status      string    `json:"status"` // "pending", "running", "completed", "failed"
	StartedAt   time.Time `json:"started_at"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
	ImageName   string    `json:"image_name,omitempty"`
}

// DeploymentConfig represents configuration for deployment
type DeploymentConfig struct {
	Image        string            `json:"image"`
	Name         string            `json:"name"`
	ProjectID    string            `json:"project_id"`
	PrivateImage bool              `json:"private_image"`
	Replicas     int               `json:"replicas"`
	Port         int               `json:"port"`
	EnvVars      map[string]string `json:"env_vars"`
}
