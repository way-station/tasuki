package config

import "testing"

// setRequired sets the required variables to valid values; a subtest then overrides one.
func setRequired(t *testing.T) {
	t.Helper()
	t.Setenv("MONGODB_URI", "mongodb://localhost:27017")
	t.Setenv("DATABASE_URL", "postgres://localhost:5432/tasuki")
	t.Setenv("FRONTEND_ORIGIN", "http://localhost:5173")
}

func TestLoad(t *testing.T) {
	t.Run("requires MONGODB_URI", func(t *testing.T) {
		setRequired(t)
		t.Setenv("MONGODB_URI", "")
		if _, err := Load(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("requires DATABASE_URL", func(t *testing.T) {
		setRequired(t)
		t.Setenv("DATABASE_URL", "")
		if _, err := Load(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("applies defaults", func(t *testing.T) {
		setRequired(t)
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
		setRequired(t)
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
		setRequired(t)
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
		setRequired(t)
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

	// ADR-0014: the Origin check compares against this value as browsers send it.
	t.Run("reads FRONTEND_ORIGIN as browsers send it", func(t *testing.T) {
		for value, want := range map[string]string{
			"https://tasuki.example":  "https://tasuki.example",
			"https://tasuki.example/": "https://tasuki.example",
			"HTTPS://Tasuki.Example":  "https://tasuki.example",
			"http://localhost:5173":   "http://localhost:5173",
		} {
			setRequired(t)
			t.Setenv("FRONTEND_ORIGIN", value)
			cfg, err := Load()
			if err != nil {
				t.Fatalf("FRONTEND_ORIGIN=%q: unexpected error: %v", value, err)
			}
			if cfg.FrontendOrigin != want {
				t.Errorf("FRONTEND_ORIGIN=%q: FrontendOrigin = %q, want %q", value, cfg.FrontendOrigin, want)
			}
		}
	})

	t.Run("refuses a FRONTEND_ORIGIN that is not an origin", func(t *testing.T) {
		for _, value := range []string{"", "tasuki.example", "ftp://tasuki.example", "https://", "https://tasuki.example/app", "https://tasuki.example?x=1", "https://user@tasuki.example"} {
			setRequired(t)
			t.Setenv("FRONTEND_ORIGIN", value)
			if _, err := Load(); err == nil {
				t.Errorf("FRONTEND_ORIGIN=%q: expected error, got nil", value)
			}
		}
	})
}
