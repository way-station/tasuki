// Package config loads backend settings from environment variables.
package config

import (
	"errors"
	"os"
)

// Config holds the runtime settings of the backend.
type Config struct {
	Port          string
	MongoURI      string
	MongoDatabase string
	DatabaseURL   string
}

// Load reads the configuration from the environment.
// MONGODB_URI and DATABASE_URL are required; the others fall back to defaults.
func Load() (Config, error) {
	cfg := Config{
		Port:          getenv("PORT", "8080"),
		MongoURI:      os.Getenv("MONGODB_URI"),
		MongoDatabase: getenv("MONGODB_DATABASE", "tasuki"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
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
