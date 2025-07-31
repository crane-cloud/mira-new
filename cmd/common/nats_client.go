package common

import (
	"encoding/json"
	"fmt"
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

// PublishBuildRequest publishes a build request to the build queue
func (c *NATSClient) PublishBuildRequest(request *BuildRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal build request: %v", err)
	}

	err = c.conn.Publish("mira.build.requests", data)
	if err != nil {
		return fmt.Errorf("failed to publish build request: %v", err)
	}

	return nil
}

// SubscribeToBuildRequests subscribes to build requests
func (c *NATSClient) SubscribeToBuildRequests(handler func(*BuildRequest)) (*nats.Subscription, error) {
	return c.conn.Subscribe("mira.build.requests", func(msg *nats.Msg) {
		var request BuildRequest
		if err := json.Unmarshal(msg.Data, &request); err != nil {
			fmt.Printf("Failed to unmarshal build request: %v\n", err)
			return
		}
		handler(&request)
	})
}

// PublishBuildStatus publishes build status updates
func (c *NATSClient) PublishBuildStatus(status *BuildStatus) error {
	data, err := json.Marshal(status)
	if err != nil {
		return fmt.Errorf("failed to marshal build status: %v", err)
	}

	subject := fmt.Sprintf("mira.status.%s", status.BuildID)
	err = c.conn.Publish(subject, data)
	if err != nil {
		return fmt.Errorf("failed to publish build status: %v", err)
	}

	return nil
}

// SubscribeToLogs subscribes to logs for a specific build
func (c *NATSClient) SubscribeToLogs(buildID string, handler func(*LogMessage)) (*nats.Subscription, error) {
	subject := fmt.Sprintf("mira.logs.%s", buildID)
	return c.conn.Subscribe(subject, func(msg *nats.Msg) {
		var logMsg LogMessage
		if err := json.Unmarshal(msg.Data, &logMsg); err != nil {
			fmt.Printf("Failed to unmarshal log message: %v\n", err)
			return
		}
		handler(&logMsg)
	})
}
