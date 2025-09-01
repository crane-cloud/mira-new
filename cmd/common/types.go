package common

import (
	"time"
)

// BuildRequest represents a containerization request sent via NATS
type BuildRequest struct {
	ID        string           `json:"id"`
	Name      string           `json:"name"`
	Spec      ImageBuilderSpec `json:"spec"`
	Timestamp time.Time        `json:"timestamp"`
}

// ImageBuilderSpec contains the build configuration
type ImageBuilderSpec struct {
	Source       ImageBuilderSource `json:"source"`
	BuildCommand string             `json:"buildCommand"`
	OutputDir    string             `json:"outputDir"`
	ProjectID    string             `json:"projectId,omitempty"`
	AccessToken  string             `json:"accessToken"`
	SSR          bool               `json:"ssr"`
	Port         int                `json:"port,omitempty"`
	Env          map[string]string  `json:"env"`
}

// ImageBuilderSource represents the source code location
type ImageBuilderSource struct {
	Type     string               `json:"sourceType"`
	GitRepo  ImageBuilderGitRepo  `json:"gitRepo,omitempty"`
	BlobFile ImageBuilderBlobFile `json:"blobFile,omitempty"`
}

// ImageBuilderGitRepo represents Git repository configuration
type ImageBuilderGitRepo struct {
	URL      string `json:"url"`
	Branch   string `json:"branch,omitempty"`
	Revision string `json:"revision,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// ImageBuilderBlobFile represents uploaded file configuration
type ImageBuilderBlobFile struct {
	Source string `json:"source"`
}

// BuildResponse represents the response from build request
type BuildResponse struct {
	BuildID string `json:"build_id"`
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message"`
	LogsURL string `json:"logs_url,omitempty"`
}

// LogMessage represents a log entry for streaming
type LogMessage struct {
	BuildID   string    `json:"build_id"`
	Level     string    `json:"level"` // info, error, debug
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
	Step      string    `json:"step,omitempty"` // clone, build, deploy, etc.
}

// BuildStatus represents the current status of a build
type BuildStatus struct {
	BuildID     string    `json:"build_id"`
	ProjectID   string    `json:"project_id,omitempty"`
	AppName     string    `json:"app_name,omitempty"`
	Status      string    `json:"status"` // pending, running, completed, failed
	StartedAt   time.Time `json:"started_at,omitempty"`
	CompletedAt time.Time `json:"completed_at,omitempty"`
	Error       string    `json:"error,omitempty"`
	ImageName   string    `json:"image_name,omitempty"`
}

// BuildCompletionMessage represents a build completion notification sent via WebSocket
type BuildCompletionMessage struct {
	Type      string    `json:"type"` // "build_completion"
	BuildID   string    `json:"build_id"`
	Status    string    `json:"status"`  // "completed" or "failed"
	Message   string    `json:"message"` // Human readable message
	Error     string    `json:"error,omitempty"`
	ImageName string    `json:"image_name,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Logger interface defines the methods that any logger must implement
type Logger interface {
	Info(message string)
	Error(message string)
	Debug(message string)
	InfoWithStep(step, message string)
	ErrorWithStep(step, message string)
	Log(fields map[string]string, message string)
	Write(p []byte) (n int, err error)
}
