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
	client *mongo.Client
	db     *mongo.Database
}

// Connect opens a client and verifies the connection with a ping.
func Connect(ctx context.Context, uri, database string) (*Client, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("connect mongo: %w", err)
	}
	s := &Client{client: client, db: client.Database(database)}
	if err := s.Ping(ctx); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return s, nil
}

// Ping checks that the primary is reachable.
func (s *Client) Ping(ctx context.Context) error {
	if err := s.client.Ping(ctx, readpref.Primary()); err != nil {
		return fmt.Errorf("ping mongo: %w", err)
	}
	return nil
}

// Database returns the application database handle.
func (s *Client) Database() *mongo.Database {
	return s.db
}

// Close disconnects the client.
func (s *Client) Close(ctx context.Context) error {
	return s.client.Disconnect(ctx)
}
