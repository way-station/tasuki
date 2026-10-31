package config

import "testing"

func TestLoad(t *testing.T) {
	t.Run("requires MONGODB_URI", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "")
		t.Setenv("DATABASE_URL", "postgres://localhost:5432/tasuki")
		if _, err := Load(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("requires DATABASE_URL", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
		t.Setenv("DATABASE_URL", "")
		if _, err := Load(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("applies defaults", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
		t.Setenv("DATABASE_URL", "postgres://localhost:5432/tasuki")
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
		if cfg.DatabaseURL != "postgres://localhost:5432/tasuki" {
			t.Errorf("DatabaseURL = %q, want %q", cfg.DatabaseURL, "postgres://localhost:5432/tasuki")
		}
	})

	t.Run("reads the optional FRONTEND_DIR", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
		t.Setenv("DATABASE_URL", "postgres://localhost:5432/tasuki")
		t.Setenv("FRONTEND_DIR", "/frontend")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.FrontendDir != "/frontend" {
			t.Errorf("FrontendDir = %q, want %q", cfg.FrontendDir, "/frontend")
		}
	})

	t.Run("reads the organizers as a comma-separated list", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
		t.Setenv("DATABASE_URL", "postgres://localhost:5432/tasuki")
		t.Setenv("ORGANIZER_LINE_USER_IDS", " U-1, ,U-2 ")

		cfg, err := Load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cfg.OrganizerLineUserIDs) != 2 || cfg.OrganizerLineUserIDs[0] != "U-1" || cfg.OrganizerLineUserIDs[1] != "U-2" {
			t.Errorf("OrganizerLineUserIDs = %q, want [U-1 U-2]", cfg.OrganizerLineUserIDs)
		}
	})

	// ADR-0009: the dev-only login is off unless the flag is exactly "true".
	t.Run("keeps the dev-only login off unless DEV_LOGIN_ENABLED is true", func(t *testing.T) {
		t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
		t.Setenv("DATABASE_URL", "postgres://localhost:5432/tasuki")
		for value, want := range map[string]bool{"": false, "false": false, "1": false, "TRUE": false, "true": true} {
			t.Setenv("DEV_LOGIN_ENABLED", value)
			cfg, err := Load()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if cfg.DevLoginEnabled != want {
				t.Errorf("DEV_LOGIN_ENABLED=%q: DevLoginEnabled = %v, want %v", value, cfg.DevLoginEnabled, want)
			}
		}
	})
}
