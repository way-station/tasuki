//go:build integration

package mongodb

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestConnect(t *testing.T) {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		t.Skip("MONGODB_URI is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	s, err := Connect(ctx, uri, "tasuki_test")
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	defer func() { _ = s.Close(ctx) }()

	if err := s.Ping(ctx); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}
