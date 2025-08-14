package common

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSLogger implements a logger that publishes logs to NATS for real-time streaming
type NATSLogger struct {
	nc      *nats.Conn
	js      nats.JetStreamContext
	buildID string
	subject string
}

// NewNATSLogger creates a new NATS-based logger
func NewNATSLogger(nc *nats.Conn, buildID string) *NATSLogger {
	js, err := nc.JetStream()
	if err != nil {
		log.Printf("Failed to get JetStream context: %v", err)
		return nil
	}

	// Ensure the log stream exists
	streamName := "MIRA_LOGS"
	_, err = js.StreamInfo(streamName)
	if err != nil {
		// Stream doesn't exist, create it
		_, err = js.AddStream(&nats.StreamConfig{
			Name:      streamName,
			Subjects:  []string{"mira.logs.*"}, // Match all log subjects
			Storage:   nats.FileStorage,
			Retention: nats.LimitsPolicy,
			MaxAge:    24 * time.Hour, // Keep logs for 24 hours
			MaxMsgs:   10000,          // Keep max 10k messages
		})
		if err != nil {
			// Check if the error is due to subject overlap
			if strings.Contains(err.Error(), "subjects overlap") {
				log.Printf("Stream with overlapping subjects already exists. Using existing stream.")
			} else {
				log.Printf("Failed to create log stream: %v", err)
			}
		}
	}

	return &NATSLogger{
		nc:      nc,
		buildID: buildID,
		subject: BuildLogsSubject(buildID),
		js:      js,
	}
}

// Log publishes a log message to NATS (compatible with Conveyor's logger interface)
func (l *NATSLogger) Log(fields map[string]string, message string) {
	l.logWithLevel("info", message, "")
}

// Info logs an info level message
func (l *NATSLogger) Info(message string) {
	l.logWithLevel("info", message, "")
}

// Error logs an error level message
func (l *NATSLogger) Error(message string) {
	l.logWithLevel("error", message, "")
}

// Debug logs a debug level message
func (l *NATSLogger) Debug(message string) {
	l.logWithLevel("debug", message, "")
}

// InfoWithStep logs an info message with a specific build step
func (l *NATSLogger) InfoWithStep(step, message string) {
	l.logWithLevel("info", message, step)
}

// ErrorWithStep logs an error message with a specific build step
func (l *NATSLogger) ErrorWithStep(step, message string) {
	l.logWithLevel("error", message, step)
}

// logWithLevel publishes a log message with the specified level
func (l *NATSLogger) logWithLevel(level, message, step string) {
	logMsg := LogMessage{
		BuildID:   l.buildID,
		Level:     level,
		Message:   message,
		Timestamp: time.Now(),
		Step:      step,
	}

	jsonData, err := json.Marshal(logMsg)
	if err != nil {
		log.Printf("Failed to marshal log message: %v", err)
		return
	}

	err = l.nc.Publish(l.subject, jsonData)
	if err != nil {
		log.Printf("Failed to publish log message: %v", err)
	}

	// Publish to JetStream for persistence
	_, err = l.js.Publish(l.subject, jsonData)
	if err != nil {
		log.Printf("Failed to publish log message to JetStream: %v", err)
		// Try to ensure stream exists and retry
		streamName := "MIRA_LOGS"
		_, streamErr := l.js.StreamInfo(streamName)
		if streamErr != nil {
			log.Printf("Stream doesn't exist, creating it...")
			_, streamErr = l.js.AddStream(&nats.StreamConfig{
				Name:      streamName,
				Subjects:  []string{"mira.logs.*"},
				Storage:   nats.FileStorage,
				Retention: nats.LimitsPolicy,
				MaxAge:    24 * time.Hour,
				MaxMsgs:   10000,
			})
			if streamErr == nil {
				// Retry publishing after creating stream
				_, err = l.js.Publish(l.subject, jsonData)
				if err != nil {
					log.Printf("Failed to publish log message to JetStream after stream creation: %v", err)
				} else {
					log.Printf("Successfully published log to JetStream after stream creation")
				}
			}
		}
	} else {
		log.Printf("Successfully published log to JetStream: %s", l.subject)
	}

	// Also log to stdout for debugging
	if step != "" {
		fmt.Printf("[%s][%s][%s] %s\n",
			logMsg.Timestamp.Format("15:04:05"), level, step, message)
	} else {
		fmt.Printf("[%s][%s] %s\n",
			logMsg.Timestamp.Format("15:04:05"), level, message)
	}
}

// Write implements io.Writer interface for compatibility with pack logger
func (l *NATSLogger) Write(p []byte) (n int, err error) {
	l.Info(string(p))
	return len(p), nil
}
