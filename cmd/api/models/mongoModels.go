package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoLogMessage represents a log entry stored in MongoDB
type MongoLogMessage struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	BuildID   string             `bson:"build_id" json:"build_id"`
	ProjectID string             `bson:"project_id,omitempty" json:"project_id,omitempty"`
	AppName   string             `bson:"app_name,omitempty" json:"app_name,omitempty"`
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
func ToMongoLogMessage(buildID, projectID, appName, level, message, step string, timestamp time.Time) MongoLogMessage {
	now := time.Now()
	return MongoLogMessage{
		BuildID:   buildID,
		ProjectID: projectID,
		AppName:   appName,
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
		ProjectID: m.ProjectID,
		AppName:   m.AppName,
		Level:     m.Level,
		Message:   m.Message,
		Timestamp: m.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
		Step:      m.Step,
	}
}
