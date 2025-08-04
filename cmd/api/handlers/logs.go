package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	common "mira/cmd/common"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// LogHandler handles WebSocket log streaming
type LogHandler struct {
	natsClient *common.NATSClient
}

// NewLogHandler creates a new log handler
func NewLogHandler(natsClient *common.NATSClient) *LogHandler {
	return &LogHandler{
		natsClient: natsClient,
	}
}

// StreamLogs handles WebSocket connections for streaming build logs
func (h *LogHandler) StreamLogs(c *websocket.Conn) {
	buildID := c.Params("buildId")
	if buildID == "" {
		log.Printf("No build ID provided for log streaming")
		c.WriteMessage(websocket.TextMessage, []byte(`{"error":"No build ID provided"}`))
		c.Close()
		return
	}

	log.Printf("Starting log stream for build ID: %s", buildID)

	// Subscribe to logs for this specific build
	sub, err := h.natsClient.SubscribeToLogs(buildID, func(logMsg *common.LogMessage) {
		// Convert log message to JSON and send via WebSocket
		data, err := json.Marshal(logMsg)
		if err != nil {
			log.Printf("Failed to marshal log message: %v", err)
			return
		}

		err = c.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Printf("Failed to write WebSocket message: %v", err)
		}
	})

	if err != nil {
		log.Printf("Failed to subscribe to logs for build %s: %v", buildID, err)
		errorMsg := fmt.Sprintf(`{"error":"Failed to subscribe to logs: %s"}`, err.Error())
		c.WriteMessage(websocket.TextMessage, []byte(errorMsg))
		c.Close()
		return
	}

	defer func() {
		if sub != nil {
			sub.Unsubscribe()
		}
		log.Printf("Log stream ended for build ID: %s", buildID)
	}()

	// Send initial connection confirmation
	confirmMsg := fmt.Sprintf(`{"message":"Connected to log stream for build %s"}`, buildID)
	c.WriteMessage(websocket.TextMessage, []byte(confirmMsg))

	// Keep connection alive and handle client messages
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error for build %s: %v", buildID, err)
			}
			break
		}

		// Handle ping/pong or other control messages
		if messageType == websocket.PingMessage {
			c.WriteMessage(websocket.PongMessage, nil)
		} else if messageType == websocket.TextMessage {
			// Handle any client commands if needed
			log.Printf("Received message from client: %s", string(message))
		}
	}
}

// WebSocketUpgrade checks if the request can be upgraded to WebSocket
func (h *LogHandler) WebSocketUpgrade(c *fiber.Ctx) error {
	// Check if it's a WebSocket upgrade request
	if websocket.IsWebSocketUpgrade(c) {
		return websocket.New(h.StreamLogs)(c)
	}
	return fiber.ErrUpgradeRequired
}
