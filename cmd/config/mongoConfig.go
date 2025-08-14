package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI      string
	Database string
	Client   *mongo.Client
}

// NewMongoDBConfig creates a new MongoDB configuration
func NewMongoDBConfig() *MongoDBConfig {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		uri = "mongodb://admin:password@localhost:27017/mira?authSource=admin"
	}

	database := os.Getenv("MONGODB_DATABASE")
	if database == "" {
		database = "mira"
	}

	return &MongoDBConfig{
		URI:      uri,
		Database: database,
	}
}

// Connect establishes a connection to MongoDB
func (c *MongoDBConfig) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(c.URI))
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Ping the database to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	c.Client = client
	log.Printf("Successfully connected to MongoDB: %s", c.URI)
	return nil
}

// Disconnect closes the MongoDB connection
func (c *MongoDBConfig) Disconnect() error {
	if c.Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return c.Client.Disconnect(ctx)
	}
	return nil
}

// GetDatabase returns the MongoDB database instance
func (c *MongoDBConfig) GetDatabase() *mongo.Database {
	if c.Client == nil {
		return nil
	}
	return c.Client.Database(c.Database)
}

// GetCollection returns a MongoDB collection
func (c *MongoDBConfig) GetCollection(name string) *mongo.Collection {
	db := c.GetDatabase()
	if db == nil {
		return nil
	}
	return db.Collection(name)
}
