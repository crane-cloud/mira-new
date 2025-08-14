package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"sort"

	"mira/cmd/api/models"
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

// GetBuildLogs retrieves logs from JetStream for a specific build ID
//
// This endpoint fetches all historical logs for a build from the JetStream storage.
// The logs are sorted by timestamp and returned as a JSON array.
//
// GET /api/logs/:buildId/history
//
// Response:
//
//	{
//	  "build_id": "build-123",
//	  "logs": [
//	    {
//	      "build_id": "build-123",
//	      "level": "info",
//	      "message": "Build started",
//	      "timestamp": "2024-01-01T12:00:00Z",
//	      "step": "clone"
//	    }
//	  ],
//	  "count": 1
//	}
//
// @Summary Get build logs history
// @Description Retrieves all historical logs for a specific build from JetStream storage
// @Tags logs
// @Accept json
// @Produce json
// @Param buildId path string true "Build ID" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} models.BuildLogsResponse "Build logs retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Build ID is required"
// @Failure 500 {object} models.ErrorResponse "Failed to retrieve logs"
// @Router /logs/{buildId}/history [get]
func (h *LogHandler) GetBuildLogs(c *fiber.Ctx) error {
	buildID := c.Params("buildId")
	if buildID == "" {
		return c.Status(400).JSON(models.ErrorResponse{
			Error: "Build ID is required",
		})
	}

	// Get logs from JetStream
	logs, err := h.natsClient.GetBuildLogs(buildID)
	if err != nil {
		log.Printf("Failed to get logs for build %s: %v", buildID, err)
		return c.Status(500).JSON(models.ErrorResponse{
			Error: "Failed to retrieve logs",
		})
	}

	// Sort logs by timestamp
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.Before(logs[j].Timestamp)
	})

	// Convert common.LogMessage to models.LogMessage
	var responseLogs []models.LogMessage
	for _, log := range logs {
		responseLogs = append(responseLogs, models.LogMessage{
			BuildID:   log.BuildID,
			Level:     log.Level,
			Message:   log.Message,
			Timestamp: log.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
			Step:      log.Step,
		})
	}

	return c.JSON(models.BuildLogsResponse{
		BuildID: buildID,
		Logs:    responseLogs,
		Count:   len(responseLogs),
	})
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
