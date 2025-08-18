package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// NATSClient provides utilities for NATS connections
type NATSClient struct {
	conn *nats.Conn
	url  string
}

// NewNATSClient creates a new NATS client
func NewNATSClient() (*NATSClient, error) {
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	conn, err := nats.Connect(natsURL,
		nats.ReconnectWait(2*time.Second),
		nats.MaxReconnects(10),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			fmt.Printf("NATS disconnected: %v\n", err)
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS reconnected to %v\n", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %v", err)
	}

	return &NATSClient{
		conn: conn,
		url:  natsURL,
	}, nil
}

// GetConnection returns the NATS connection
func (c *NATSClient) GetConnection() *nats.Conn {
	return c.conn
}

// GetJetStream returns the JetStream context
//
// This method provides access to the JetStream context for advanced
// operations like stream management, consumer creation, and message
// retrieval from persistent storage.
func (c *NATSClient) GetJetStream() (nats.JetStreamContext, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("NATS connection is not healthy")
	}
	return c.conn.JetStream()
}

// Close closes the NATS connection
func (c *NATSClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}

// IsConnected checks if the NATS connection is healthy
func (c *NATSClient) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

// GetStats returns connection statistics
func (c *NATSClient) GetStats() nats.Statistics {
	if c.conn == nil {
		return nats.Statistics{}
	}
	return c.conn.Stats()
}

// PublishBuildRequest publishes a build request to the build queue with enhanced error handling
func (c *NATSClient) PublishBuildRequest(request *BuildRequest) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.PublishBuildRequestWithContext(ctx, request)
}

// PublishBuildRequestWithContext publishes a build request with context support and retry logic
func (c *NATSClient) PublishBuildRequestWithContext(ctx context.Context, request *BuildRequest) error {
	if !c.IsConnected() {
		return fmt.Errorf("NATS connection is not healthy")
	}

	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal build request: %v", err)
	}

	// Retry logic with exponential backoff
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("publish cancelled: %v", ctx.Err())
		default:
		}

		err = c.conn.Publish(SUBJECT_BUILD_REQUEST, data)
		if err == nil {
			log.Printf("Build request %s published successfully on attempt %d", request.ID, attempt+1)
			return nil
		}

		if attempt == maxRetries-1 {
			break // Last attempt failed, don't wait
		}

		// Exponential backoff
		delay := baseDelay * time.Duration(1<<attempt)
		log.Printf("Publish attempt %d failed, retrying in %v: %v", attempt+1, delay, err)

		select {
		case <-ctx.Done():
			return fmt.Errorf("publish cancelled during retry: %v", ctx.Err())
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("failed to publish build request after %d attempts: %v", maxRetries, err)
}

// PublishBuildRequestAsync publishes a build request asynchronously with callback
func (c *NATSClient) PublishBuildRequestAsync(request *BuildRequest, onSuccess func(), onError func(error)) {
	go func() {
		err := c.PublishBuildRequest(request)
		if err != nil {
			log.Printf("Async publish failed for build %s: %v", request.ID, err)
			if onError != nil {
				onError(err)
			}
		} else {
			log.Printf("Async publish succeeded for build %s", request.ID)
			if onSuccess != nil {
				onSuccess()
			}
		}
	}()
}

// SubscribeToBuildRequests subscribes to build requests
func (c *NATSClient) SubscribeToBuildRequests(handler func(*BuildRequest)) (*nats.Subscription, error) {
	return c.conn.Subscribe(SUBJECT_BUILD_REQUEST, func(msg *nats.Msg) {
		var request BuildRequest
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			fmt.Printf("Failed to unmarshal build request: %v\n", err)
			return
		}
		handler(&request)
	})
}

// PublishBuildStatus publishes build status updates with enhanced error handling
func (c *NATSClient) PublishBuildStatus(status *BuildStatus) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.PublishBuildStatusWithContext(ctx, status)
}

// PublishBuildStatusWithContext publishes build status with context support and retry logic
func (c *NATSClient) PublishBuildStatusWithContext(ctx context.Context, status *BuildStatus) error {
	if !c.IsConnected() {
		return fmt.Errorf("NATS connection is not healthy")
	}

	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal build status: %v", err)
	}

	subject := BuildStatusSubject(status.BuildID)

	// Retry logic with exponential backoff (fewer retries for status updates)
	maxRetries := 2
	baseDelay := 50 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("status publish cancelled: %v", ctx.Err())
		default:
		}

		err = c.conn.Publish(subject, data)
		if err == nil {
			return nil
		}

		if attempt == maxRetries-1 {
			break // Last attempt failed, don't wait
		}

		// Exponential backoff
		delay := baseDelay * time.Duration(1<<attempt)

		select {
		case <-ctx.Done():
			return fmt.Errorf("status publish cancelled during retry: %v", ctx.Err())
		case <-time.After(delay):
		}
	}

	return fmt.Errorf("failed to publish build status after %d attempts: %v", maxRetries, err)
}

// SubscribeToLogs subscribes to logs for a specific build
func (c *NATSClient) SubscribeToLogs(buildID string, handler func(*LogMessage)) (*nats.Subscription, error) {
	subject := BuildLogsSubject(buildID)
	return c.conn.Subscribe(subject, func(msg *nats.Msg) {
		var logMsg LogMessage
		if err := json.Unmarshal(msg.Data, &logMsg); err != nil {
			fmt.Printf("Failed to unmarshal log message: %v\n", err)
			return
		}
		handler(&logMsg)
	})
}

// GetBuildLogs retrieves all logs for a specific build from JetStream
//
// This method fetches all historical log messages for a specific build ID
// from the JetStream storage. It creates a temporary consumer to read
// all messages from the build's log subject and returns them as an array.
//
// The method automatically ensures the log stream exists and cleans up
// the temporary consumer after use.
func (c *NATSClient) GetBuildLogs(buildID string) ([]LogMessage, error) {
	js, err := c.GetJetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %v", err)
	}

	subject := BuildLogsSubject(buildID)
	streamName := "MIRA_LOGS" // Default stream name

	// Try to ensure the stream exists, but don't fail if it doesn't
	err = c.ensureLogStream(js, streamName)
	if err != nil {
		log.Printf("Warning: Failed to ensure log stream: %v", err)
		// Continue anyway, we'll try to find existing streams
	}

	// Create a unique durable consumer to ensure DeliverAll per request
	consumerName := fmt.Sprintf("hist_%s_%d", buildID, time.Now().UnixNano())
	_, err = js.AddConsumer(streamName, &nats.ConsumerConfig{
		Durable:           consumerName,
		FilterSubject:     subject,
		AckPolicy:         nats.AckExplicitPolicy,
		DeliverPolicy:     nats.DeliverAllPolicy,
		InactiveThreshold: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add consumer: %v", err)
	}
	defer js.DeleteConsumer(streamName, consumerName)

	// Subscribe to the durable consumer bound to the stream
	sub, err := js.PullSubscribe(subject, consumerName, nats.BindStream(streamName))
	if err != nil {
		return nil, fmt.Errorf("failed to create subscription: %v", err)
	}
	defer sub.Unsubscribe()

	// Fetch all available messages
	var logs []LogMessage
	batchSize := 100

	for {
		messages, err := sub.Fetch(batchSize, nats.MaxWait(5*time.Second))
		if err != nil {
			if err == nats.ErrTimeout {
				break // No more messages
			}
			return nil, fmt.Errorf("failed to fetch messages: %v", err)
		}

		if len(messages) == 0 {
			break // No more messages
		}

		for _, msg := range messages {
			var logMsg LogMessage
			if err := json.Unmarshal(msg.Data, &logMsg); err != nil {
				msg.Ack()
				continue
			}

			logs = append(logs, logMsg)
			msg.Ack()
		}

		if len(messages) < batchSize {
			break // No more messages to fetch
		}
	}

	// If we found logs, return them
	if len(logs) > 0 {
		return logs, nil
	}

	// If no logs found, return empty array (this is normal for new builds)
	log.Printf("No logs found for build %s", buildID)
	return logs, nil
}

// ensureLogStream ensures that the log stream exists, creating it if necessary
func (c *NATSClient) ensureLogStream(js nats.JetStreamContext, streamName string) error {
	// Try to get stream info to check if it exists
	streamInfo, err := js.StreamInfo(streamName)
	if err == nil {
		// Stream exists, check if it has the correct subjects
		hasCorrectSubjects := false
		for _, subject := range streamInfo.Config.Subjects {
			if subject == "mira.logs.*" {
				hasCorrectSubjects = true
				break
			}
		}

		if hasCorrectSubjects {
			return nil // Stream exists with correct configuration
		}

		// Stream exists but doesn't have the correct subjects
		// We can't modify subjects of an existing stream, so we'll use it as is
		log.Printf("Stream %s exists but doesn't have mira.logs.* subject. Using existing stream.", streamName)
		return nil
	}

	// Configurable retention
	maxAgeHours := 24
	if v := os.Getenv("MIRA_LOG_STREAM_MAX_AGE_HOURS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			maxAgeHours = n
		}
	}
	maxMsgs := 10000
	if v := os.Getenv("MIRA_LOG_STREAM_MAX_MSGS"); v != "" {
		if n, e := strconv.Atoi(v); e == nil && n > 0 {
			maxMsgs = n
		}
	}

	// Stream doesn't exist, create it
	_, err = js.AddStream(&nats.StreamConfig{
		Name:      streamName,
		Subjects:  []string{"mira.logs.*"}, // Match all log subjects
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    time.Duration(maxAgeHours) * time.Hour,
		MaxMsgs:   int64(maxMsgs),
	})
	if err != nil {
		// Check if the error is due to subject overlap
		if strings.Contains(err.Error(), "subjects overlap") {
			log.Printf("Stream with overlapping subjects already exists. Using existing stream.")
			return nil
		}
		return fmt.Errorf("failed to create log stream: %v", err)
	}

	return nil
}
