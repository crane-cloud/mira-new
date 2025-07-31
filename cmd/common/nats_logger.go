package common

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSLogger implements a logger that publishes logs to NATS for real-time streaming
type NATSLogger struct {
	nc      *nats.Conn
	buildID string
	subject string
}

// NewNATSLogger creates a new NATS-based logger
func NewNATSLogger(nc *nats.Conn, buildID string) *NATSLogger {
	return &NATSLogger{
		nc:      nc,
		buildID: buildID,
		subject: fmt.Sprintf("mira.logs.%s", buildID),
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

	data, err := json.Marshal(logMsg)
	if err != nil {
		log.Printf("Failed to marshal log message: %v", err)
		return
	}

	err = l.nc.Publish(l.subject, data)
	if err != nil {
		log.Printf("Failed to publish log message: %v", err)
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
