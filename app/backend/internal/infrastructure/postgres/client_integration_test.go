//go:build integration

package postgres

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	raw := os.Getenv("DATABASE_URL")
	if raw == "" {
		t.Skip("DATABASE_URL is not set")
	}
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}
	u.Path = "/tasuki_test"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := Connect(ctx, u.String())
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer c.Close()

	if err := c.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	var db string
	if err := c.Pool().QueryRow(ctx, "SELECT current_database()").Scan(&db); err != nil {
		t.Fatalf("current_database: %v", err)
	}
	if db != "tasuki_test" {
		t.Errorf("current_database() = %q, want %q", db, "tasuki_test")
	}
}
