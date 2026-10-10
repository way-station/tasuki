package config

import "testing"

func TestLoad(t *testing.T) {
	t.Run("requires MONGODB_URI", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "")
		if _, err := Load(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("applies defaults", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
		t.Setenv("PORT", "")
		t.Setenv("MONGODB_DATABASE", "")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.Port != "8080" {
			t.Errorf("Port = %q, want %q", cfg.Port, "8080")
		}
		if cfg.MongoDatabase != "tasuki" {
			t.Errorf("MongoDatabase = %q, want %q", cfg.MongoDatabase, "tasuki")
		}
	})
}
