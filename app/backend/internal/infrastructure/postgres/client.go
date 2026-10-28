// Package postgres holds the PostgreSQL connection pool and the PostgreSQL
// implementations of the repositories and query services.
package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Client wraps the PostgreSQL connection pool.
type Client struct {
	pool *pgxpool.Pool
}

// Connect opens a connection pool and verifies the connection with a ping.
// Pool sizes are read from the connection string (pool_max_conns and others, ADR-0015).
func Connect(ctx context.Context, url string) (*Client, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	client := &Client{pool: pool}
	if err := client.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return client, nil
}

// Ping checks that the database is reachable.
func (client *Client) Ping(ctx context.Context) error {
	if err := client.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}
	return nil
}

// Pool returns the connection pool.
func (client *Client) Pool() *pgxpool.Pool {
	return client.pool
}

// Close closes every connection in the pool. It waits for connections in use to be released.
func (client *Client) Close() {
	client.pool.Close()
}
