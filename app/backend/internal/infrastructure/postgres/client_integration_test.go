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
	databaseURL, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse DATABASE_URL: %v", err)
	}
	databaseURL.Path = "/tasuki_test"

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := Connect(ctx, databaseURL.String())
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer client.Close()

	if err := client.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
	var databaseName string
	if err := client.Pool().QueryRow(ctx, "SELECT current_database()").Scan(&databaseName); err != nil {
		t.Fatalf("current_database: %v", err)
	}
	if databaseName != "tasuki_test" {
		t.Errorf("current_database() = %q, want %q", databaseName, "tasuki_test")
	}
}
