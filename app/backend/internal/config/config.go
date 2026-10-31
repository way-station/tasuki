// Package config loads backend settings from environment variables.
package config

import (
	"errors"
	"os"
	"strings"
)

// Config holds the runtime settings of the backend.
type Config struct {
	Port          string
	MongoURI      string
	MongoDatabase string
	DatabaseURL   string
	// FrontendDir holds the built frontend to serve. Empty in development,
	// where the Vite dev server serves the frontend.
	FrontendDir string
	// OrganizerLineUserIDs lists the organizers by LINE user ID (ADR-0006).
	OrganizerLineUserIDs []string
	// DevLoginEnabled turns on the dev-only login (ADR-0009). Set only in development.
	DevLoginEnabled bool
}

// Load reads the configuration from the environment.
// MONGODB_URI and DATABASE_URL are required. FRONTEND_DIR, ORGANIZER_LINE_USER_IDS
// (comma-separated), and DEV_LOGIN_ENABLED (only "true" turns it on) are optional. The
// others fall back to defaults.
func Load() (Config, error) {
	cfg := Config{
		Port:          getenv("PORT", "8080"),
		MongoURI:      os.Getenv("MONGODB_URI"),
		MongoDatabase: getenv("MONGODB_DATABASE", "tasuki"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		FrontendDir:   os.Getenv("FRONTEND_DIR"),

		OrganizerLineUserIDs: splitList(os.Getenv("ORGANIZER_LINE_USER_IDS")),
		DevLoginEnabled:      os.Getenv("DEV_LOGIN_ENABLED") == "true",
	}
	if cfg.MongoURI == "" {
		return Config{}, errors.New("MONGODB_URI is required")
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	return cfg, nil
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// splitList splits a comma-separated list, dropping blanks.
func splitList(value string) []string {
	var items []string
	for item := range strings.SplitSeq(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}
