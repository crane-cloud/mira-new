package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"mira/cmd/api/models"
	"mira/cmd/config"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoLogService handles MongoDB operations for logs
type MongoLogService struct {
	mongoConfig *config.MongoDBConfig
	collection  *mongo.Collection
}

// NewMongoLogService creates a new MongoDB log service
func NewMongoLogService(mongoConfig *config.MongoDBConfig) *MongoLogService {
	logsCollection := mongoConfig.GetCollection("logs")
	buildsCollection := mongoConfig.GetCollection("builds")

	// Create indexes for better query performance
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Create compound index on build_id and timestamp
		indexModel := mongo.IndexModel{
			Keys: bson.D{
				{Key: "build_id", Value: 1},
				{Key: "timestamp", Value: 1},
			},
			Options: options.Index().SetName("build_id_timestamp_idx"),
		}

		// Create all indexes for logs collection
		logsIndexes := []mongo.IndexModel{indexModel}
		for _, idx := range logsIndexes {
			_, err := logsCollection.Indexes().CreateOne(ctx, idx)
			if err != nil {
				log.Printf("Failed to create index %s: %v", idx.Options.Name, err)
			} else {
				log.Printf("Successfully created index %s for logs collection", idx.Options.Name)
			}
		}

		// Create indexes for builds collection
		buildsProjectIndex := mongo.IndexModel{
			Keys: bson.D{
				{Key: "project_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("builds_project_id_created_idx"),
		}

		buildsAppNameIndex := mongo.IndexModel{
			Keys: bson.D{
				{Key: "app_name", Value: 1},
				{Key: "created_at", Value: -1},
			},
			Options: options.Index().SetName("builds_app_name_created_idx"),
		}

		buildsIndexes := []mongo.IndexModel{buildsProjectIndex, buildsAppNameIndex}
		for _, idx := range buildsIndexes {
			_, err := buildsCollection.Indexes().CreateOne(ctx, idx)
			if err != nil {
				log.Printf("Failed to create index %s: %v", idx.Options.Name, err)
			} else {
				log.Printf("Successfully created index %s for builds collection", idx.Options.Name)
			}
		}
	}()

	return &MongoLogService{
		mongoConfig: mongoConfig,
		collection:  logsCollection,
	}
}

// SaveLog saves a single log message to MongoDB
func (s *MongoLogService) SaveLog(buildID, level, message, step string, timestamp time.Time) error {
	if s.collection == nil {
		return fmt.Errorf("MongoDB collection is not available")
	}

	mongoLog := models.ToMongoLogMessage(buildID, level, message, step, timestamp)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.collection.InsertOne(ctx, mongoLog)
	if err != nil {
		return fmt.Errorf("failed to insert log: %v", err)
	}

	return nil
}

// GetLogsByBuildID retrieves all logs for a specific build ID
func (s *MongoLogService) GetLogsByBuildID(buildID string) ([]models.LogMessage, error) {
	if s.collection == nil {
		return nil, fmt.Errorf("MongoDB collection is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Find all logs for the build ID, sorted by timestamp
	filter := bson.M{"build_id": buildID}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}}) // Sort by timestamp ascending (oldest first)

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find logs: %v", err)
	}
	defer cursor.Close(ctx)

	var mongoLogs []models.MongoLogMessage
	if err = cursor.All(ctx, &mongoLogs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %v", err)
	}

	// Convert to common LogMessage format
	var logs []models.LogMessage
	for _, mongoLog := range mongoLogs {
		logs = append(logs, mongoLog.ToLogMessage())
	}

	return logs, nil
}

// GetLogsWithFilters retrieves logs with various filters and pagination
func (s *MongoLogService) GetLogsWithFilters(buildID, level, step string, startDate, endDate *time.Time, page, limit int, sortOrder string) ([]models.LogMessage, int64, error) {
	if s.collection == nil {
		return nil, 0, fmt.Errorf("MongoDB collection is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter based on provided parameters
	filter := bson.M{}

	if buildID != "" {
		filter["build_id"] = buildID
	}
	if level != "" {
		filter["level"] = level
	}
	if step != "" {
		filter["step"] = step
	}
	if startDate != nil || endDate != nil {
		dateFilter := bson.M{}
		if startDate != nil {
			dateFilter["$gte"] = *startDate
		}
		if endDate != nil {
			dateFilter["$lte"] = *endDate
		}
		filter["timestamp"] = dateFilter
	}

	// Count total documents
	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count logs: %v", err)
	}

	// Calculate skip value for pagination
	skip := int64((page - 1) * limit)

	// Determine sort order
	sortValue := 1 // Default to ascending (oldest first)
	if sortOrder == "desc" {
		sortValue = -1 // Descending (newest first)
	}

	// Find logs with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: sortValue}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find logs: %v", err)
	}
	defer cursor.Close(ctx)

	var mongoLogs []models.MongoLogMessage
	if err = cursor.All(ctx, &mongoLogs); err != nil {
		return nil, 0, fmt.Errorf("failed to decode logs: %v", err)
	}

	// Convert to common LogMessage format
	var logs []models.LogMessage
	for _, mongoLog := range mongoLogs {
		logs = append(logs, mongoLog.ToLogMessage())
	}

	return logs, total, nil
}

// GetLogsByBuildIDWithPagination retrieves logs with pagination (backward compatibility)
func (s *MongoLogService) GetLogsByBuildIDWithPagination(buildID string, page, limit int) ([]models.LogMessage, int64, error) {
	return s.GetLogsWithFilters(buildID, "", "", nil, nil, page, limit, "asc")
}

// DeleteLogsByBuildID deletes all logs for a specific build ID
func (s *MongoLogService) DeleteLogsByBuildID(buildID string) error {
	if s.collection == nil {
		return fmt.Errorf("MongoDB collection is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"build_id": buildID}
	result, err := s.collection.DeleteMany(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete logs: %v", err)
	}

	log.Printf("Deleted %d logs for build %s", result.DeletedCount, buildID)
	return nil
}

// GetLogsByDateRange retrieves logs within a date range
func (s *MongoLogService) GetLogsByDateRange(startDate, endDate time.Time) ([]models.LogMessage, error) {
	if s.collection == nil {
		return nil, fmt.Errorf("MongoDB collection is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"timestamp": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}}) // Sort by timestamp ascending (oldest first)

	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find logs: %v", err)
	}
	defer cursor.Close(ctx)

	var mongoLogs []models.MongoLogMessage
	if err = cursor.All(ctx, &mongoLogs); err != nil {
		return nil, fmt.Errorf("failed to decode logs: %v", err)
	}

	// Convert to common LogMessage format
	var logs []models.LogMessage
	for _, mongoLog := range mongoLogs {
		logs = append(logs, mongoLog.ToLogMessage())
	}

	return logs, nil
}

// GetLogStats retrieves statistics about logs
func (s *MongoLogService) GetLogStats() (map[string]interface{}, error) {
	if s.collection == nil {
		return nil, fmt.Errorf("MongoDB collection is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Total log count
	totalLogs, err := s.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count total logs: %v", err)
	}

	// Unique build count
	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$build_id"}},
		{"$count": "unique_builds"},
	}

	cursor, err := s.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate build count: %v", err)
	}
	defer cursor.Close(ctx)

	var buildCountResult []bson.M
	if err = cursor.All(ctx, &buildCountResult); err != nil {
		return nil, fmt.Errorf("failed to decode build count: %v", err)
	}

	uniqueBuilds := int64(0)
	if len(buildCountResult) > 0 {
		if count, ok := buildCountResult[0]["unique_builds"].(int32); ok {
			uniqueBuilds = int64(count)
		}
	}

	// Logs by level
	levelPipeline := []bson.M{
		{"$group": bson.M{"_id": "$level", "count": bson.M{"$sum": 1}}},
		{"$sort": bson.M{"count": -1}},
	}

	levelCursor, err := s.collection.Aggregate(ctx, levelPipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to aggregate level stats: %v", err)
	}
	defer levelCursor.Close(ctx)

	var levelStats []bson.M
	if err = levelCursor.All(ctx, &levelStats); err != nil {
		return nil, fmt.Errorf("failed to decode level stats: %v", err)
	}

	stats := map[string]interface{}{
		"total_logs":    totalLogs,
		"unique_builds": uniqueBuilds,
		"logs_by_level": levelStats,
		"generated_at":  time.Now(),
	}

	return stats, nil
}

// SaveBuildStatus saves a build status to MongoDB
func (s *MongoLogService) SaveBuildStatus(buildID, projectID, appName, status string, startedAt, completedAt time.Time, error, imageName string) error {
	buildsCollection := s.mongoConfig.GetCollection("builds")
	if buildsCollection == nil {
		return fmt.Errorf("MongoDB builds collection is not available")
	}

	mongoBuildStatus := models.ToMongoBuildStatus(buildID, projectID, appName, status, startedAt, completedAt, error, imageName)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Use upsert to update existing build status or create new one
	filter := bson.M{"build_id": buildID}
	update := bson.M{"$set": mongoBuildStatus}
	opts := options.Update().SetUpsert(true)

	_, err := buildsCollection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("failed to save build status: %v", err)
	}

	return nil
}

// GetBuildsWithFilters retrieves builds with various filters and pagination
func (s *MongoLogService) GetBuildsWithFilters(projectID, appName, status string, page, limit int, sortOrder string) ([]models.BuildStatusResponse, int64, error) {
	buildsCollection := s.mongoConfig.GetCollection("builds")
	if buildsCollection == nil {
		return nil, 0, fmt.Errorf("MongoDB builds collection is not available")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Build filter based on provided parameters
	filter := bson.M{}

	if projectID != "" {
		filter["project_id"] = projectID
	}
	if appName != "" {
		filter["app_name"] = appName
	}
	if status != "" {
		filter["status"] = status
	}

	// Count total documents
	total, err := buildsCollection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count builds: %v", err)
	}

	// Calculate skip value for pagination
	skip := int64((page - 1) * limit)

	// Determine sort order
	sortValue := -1 // Default to descending (newest first)
	if sortOrder == "asc" {
		sortValue = 1 // Ascending (oldest first)
	}

	// Find builds with pagination
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: sortValue}}).
		SetSkip(skip).
		SetLimit(int64(limit))

	cursor, err := buildsCollection.Find(ctx, filter, opts)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find builds: %v", err)
	}
	defer cursor.Close(ctx)

	var mongoBuilds []models.MongoBuildStatus
	if err = cursor.All(ctx, &mongoBuilds); err != nil {
		return nil, 0, fmt.Errorf("failed to decode builds: %v", err)
	}

	// Convert to response format
	var builds []models.BuildStatusResponse
	for _, mongoBuild := range mongoBuilds {
		builds = append(builds, mongoBuild.ToBuildStatusResponse())
	}

	return builds, total, nil
}
