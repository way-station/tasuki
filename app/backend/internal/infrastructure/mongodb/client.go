// Package mongodb holds the MongoDB connection and the MongoDB implementations of
// the repositories and query services.
package mongodb

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
)

// Client wraps the MongoDB client and the application database.
type Client struct {
	mongoClient *mongo.Client
	database    *mongo.Database
}

// Connect opens a client and verifies the connection with a ping.
func Connect(ctx context.Context, uri, database string) (*Client, error) {
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	client := &Client{mongoClient: mongoClient, database: mongoClient.Database(database)}
	if err := client.Ping(ctx); err != nil {
		_ = mongoClient.Disconnect(ctx)
		return nil, err
	}
	return client, nil
}

// Ping checks that the primary is reachable.
func (client *Client) Ping(ctx context.Context) error {
	if err := client.mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("ping mongo: %w", err)
	}
	return nil
}

// Database returns the application database handle.
func (client *Client) Database() *mongo.Database {
	return client.database
}

// Close disconnects the client.
func (client *Client) Close(ctx context.Context) error {
	return client.mongoClient.Disconnect(ctx)
}
