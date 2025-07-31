package common

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
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
