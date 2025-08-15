package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoLogMessage represents a log entry stored in MongoDB
type MongoLogMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	BuildID   string             `bson:"build_id" json:"build_id"`
	Level     string             `bson:"level" json:"level"`
	Message   string             `bson:"message" json:"message"`
	Timestamp time.Time          `bson:"timestamp" json:"timestamp"`
	Step      string             `bson:"step,omitempty" json:"step,omitempty"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// MongoBuildLog represents a build with its associated logs
type MongoBuildLog struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	BuildID   string             `bson:"build_id" json:"build_id"`
	Logs      []MongoLogMessage  `bson:"logs" json:"logs"`
	LogCount  int                `bson:"log_count" json:"log_count"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at" json:"updated_at"`
}

// MongoLogIndex represents the indexes for the logs collection
type MongoLogIndex struct {
	BuildID   string    `bson:"build_id"`
	Timestamp time.Time `bson:"timestamp"`
}

// ToMongoLogMessage converts a common LogMessage to MongoLogMessage
func ToMongoLogMessage(buildID, level, message, step string, timestamp time.Time) MongoLogMessage {
	now := time.Now()
	return MongoLogMessage{
		BuildID:   buildID,
		Level:     level,
		Message:   message,
		Timestamp: timestamp,
		Step:      step,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// ToLogMessage converts MongoLogMessage back to common LogMessage
func (m MongoLogMessage) ToLogMessage() LogMessage {
	return LogMessage{
		BuildID:   m.BuildID,
		Level:     m.Level,
		Message:   m.Message,
		Timestamp: m.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Step:      m.Step,
	}
}

// MongoBuildStatus represents a build status stored in MongoDB
type MongoBuildStatus struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	BuildID     string             `bson:"build_id" json:"build_id"`
	ProjectID   string             `bson:"project_id,omitempty" json:"project_id,omitempty"`
	AppName     string             `bson:"app_name,omitempty" json:"app_name,omitempty"`
	Status      string             `bson:"status" json:"status"`
	StartedAt   time.Time          `bson:"started_at,omitempty" json:"started_at,omitempty"`
	CompletedAt time.Time          `bson:"completed_at,omitempty" json:"completed_at,omitempty"`
	Error       string             `bson:"error,omitempty" json:"error,omitempty"`
	ImageName   string             `bson:"image_name,omitempty" json:"image_name,omitempty"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
}

// ToBuildStatusResponse converts MongoBuildStatus to BuildStatusResponse
func (m MongoBuildStatus) ToBuildStatusResponse() BuildStatusResponse {
	response := BuildStatusResponse{
		BuildID:   m.BuildID,
		ProjectID: m.ProjectID,
		AppName:   m.AppName,
		Status:    m.Status,
		Error:     m.Error,
		ImageName: m.ImageName,
	}

	if !m.StartedAt.IsZero() {
		response.StartedAt = m.StartedAt.Format("2006-01-02T15:04:05Z07:00")
	}
	if !m.CompletedAt.IsZero() {
		response.CompletedAt = m.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
	}

	return response
}

// ToMongoBuildStatus converts a common BuildStatus to MongoBuildStatus
func ToMongoBuildStatus(buildID, projectID, appName, status string, startedAt, completedAt time.Time, error, imageName string) MongoBuildStatus {
	now := time.Now()
	return MongoBuildStatus{
		BuildID:     buildID,
		ProjectID:   projectID,
		AppName:     appName,
		Status:      status,
		StartedAt:   startedAt,
		CompletedAt: completedAt,
		Error:       error,
		ImageName:   imageName,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}
