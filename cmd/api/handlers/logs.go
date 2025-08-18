package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strconv"
	"time"

	"mira/cmd/api/models"
	"mira/cmd/api/services"
	common "mira/cmd/common"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// LogHandler handles WebSocket log streaming and MongoDB log operations
type LogHandler struct {
	natsClient   *common.NATSClient
	mongoService *services.MongoLogService
}

// NewLogHandler creates a new log handler
func NewLogHandler(natsClient *common.NATSClient, mongoService *services.MongoLogService) *LogHandler {
	return &LogHandler{
		natsClient:   natsClient,
		mongoService: mongoService,
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
func (h *LogHandler) GetJetStreamBuildLogs(c *fiber.Ctx) error {
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

// GetBuildLogsFromMongoDB retrieves logs from MongoDB with query filters
// @Summary Get build logs from MongoDB
// @Description Retrieves logs from MongoDB storage with optional filters
// @Tags logs
// @Accept json
// @Produce json
// @Param buildId query string false "Build ID filter" example("550e8400-e29b-41d4-a716-446655440000")
// @Param level query string false "Log level filter (info, error, debug)" example("info")
// @Param step query string false "Build step filter" example("clone")
// @Param startDate query string false "Start date filter (ISO 8601 format)" example("2024-01-01T00:00:00Z")
// @Param endDate query string false "End date filter (ISO 8601 format)" example("2024-01-31T23:59:59Z")
// @Param sort query string false "Sort order (asc for oldest first, desc for newest first)" example("asc")
// @Param page query int false "Page number (default: 1)" example(1)
// @Param limit query int false "Number of logs per page (default: 100, max: 1000)" example(100)
// @Success 200 {object} models.BuildLogsResponse "Build logs retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} models.ErrorResponse "Failed to retrieve logs"
// @Router /logs [get]
func (h *LogHandler) GetBuildLogsFromMongoDB(c *fiber.Ctx) error {
	// Get query parameters
	buildID := c.Query("buildId")
	level := c.Query("level")
	step := c.Query("step")
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	sortOrder := c.Query("sort", "asc") // Default to ascending (oldest first)

	if h.mongoService == nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Error: "MongoDB service is not available",
		})
	}

	// Get pagination parameters
	pageStr := c.Query("page", "1")
	limitStr := c.Query("limit", "100")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	// Parse date filters
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			startDate = &parsed
		} else {
			return c.Status(400).JSON(models.ErrorResponse{
				Error: "Invalid startDate format. Use ISO 8601 format (e.g., 2024-01-01T00:00:00Z)",
			})
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			endDate = &parsed
		} else {
			return c.Status(400).JSON(models.ErrorResponse{
				Error: "Invalid endDate format. Use ISO 8601 format (e.g., 2024-01-31T23:59:59Z)",
			})
		}
	}

	// Get logs from MongoDB with filters
	logs, total, err := h.mongoService.GetLogsWithFilters(buildID, level, step, startDate, endDate, page, limit, sortOrder)
	if err != nil {
		log.Printf("Failed to get logs from MongoDB: %v", err)
		return c.Status(500).JSON(models.ErrorResponse{
			Error: "Failed to retrieve logs from MongoDB",
		})
	}

	// Build response with filters applied
	response := fiber.Map{
		"logs":  logs,
		"count": len(logs),
		"total": total,
		"page":  page,
		"limit": limit,
		"pages": (total + int64(limit) - 1) / int64(limit), // Calculate total pages
	}

	// Add filters to response if they were applied
	if buildID != "" {
		response["build_id"] = buildID
	}
	if level != "" {
		response["level"] = level
	}
	if step != "" {
		response["step"] = step
	}
	if startDate != nil {
		response["start_date"] = startDate.Format(time.RFC3339)
	}
	if endDate != nil {
		response["end_date"] = endDate.Format(time.RFC3339)
	}
	response["sort"] = sortOrder

	return c.JSON(response)
}

// GetBuilds retrieves builds with optional filters
// @Summary Get builds with filters
// @Description Retrieves builds from MongoDB storage with optional filters
// @Tags builds
// @Accept json
// @Produce json
// @Param projectId query string false "Project ID filter" example("proj-123")
// @Param appName query string false "App name filter" example("my-app")
// @Param status query string false "Build status filter (pending, running, completed, failed)" example("completed")
// @Param sort query string false "Sort order (desc for newest first, asc for oldest first)" example("desc")
// @Param page query int false "Page number (default: 1)" example(1)
// @Param limit query int false "Number of builds per page (default: 10, max: 100)" example(10)
// @Success 200 {object} models.BuildsResponse "Builds retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} models.ErrorResponse "Failed to retrieve builds"
// @Router /builds [get]
func (h *LogHandler) GetBuilds(c *fiber.Ctx) error {
	// Get query parameters
	projectID := c.Query("projectId")
	appName := c.Query("appName")
	status := c.Query("status")
	sortOrder := c.Query("sort", "desc") // Default to descending (newest first)

	// Parse pagination parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10 // Default limit for builds
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			if l > 100 {
				l = 100 // Max limit
			}
			limit = l
		}
	}

	if h.mongoService == nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Error: "MongoDB service not available",
		})
	}

	// Get builds from MongoDB with filters
	builds, total, err := h.mongoService.GetBuildsWithFilters(projectID, appName, status, page, limit, sortOrder)
	if err != nil {
		log.Printf("Failed to get builds from MongoDB: %v", err)
		return c.Status(500).JSON(models.ErrorResponse{
			Error: "Failed to retrieve builds",
		})
	}

	// Calculate pagination info
	pages := int(math.Ceil(float64(total) / float64(limit)))

	// Build response
	response := fiber.Map{
		"builds": builds,
		"count":  len(builds),
		"total":  total,
		"page":   page,
		"limit":  limit,
		"pages":  pages,
		"sort":   sortOrder,
	}

	// Add filters to response if they were applied
	if projectID != "" {
		response["project_id"] = projectID
	}
	if appName != "" {
		response["app_name"] = appName
	}
	if status != "" {
		response["status"] = status
	}

	return c.JSON(response)
}

// GetLogStats retrieves log statistics from MongoDB
// @Summary Get log statistics
// @Description Retrieves statistics about logs stored in MongoDB
// @Tags logs
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Log statistics"
// @Failure 500 {object} models.ErrorResponse "Failed to retrieve statistics"
// @Router /logs/stats [get]
func (h *LogHandler) GetLogStats(c *fiber.Ctx) error {
	if h.mongoService == nil {
		return c.Status(500).JSON(models.ErrorResponse{
			Error: "MongoDB service is not available",
		})
	}

	stats, err := h.mongoService.GetLogStats()
	if err != nil {
		log.Printf("Failed to get log stats: %v", err)
		return c.Status(500).JSON(models.ErrorResponse{
			Error: "Failed to retrieve log statistics",
		})
	}

	return c.JSON(stats)
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
