// Command seed creates the development team's row, which the server needs to start
// (docs/design/api.md). It reads DATABASE_URL and TEAM_ID; run it with make db-seed.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "seed:", err)
		os.Exit(1)
	}
}

func run() error {
	databaseURL, teamID := os.Getenv("DATABASE_URL"), os.Getenv("TEAM_ID")
	if databaseURL == "" || teamID == "" {
		return errors.New("DATABASE_URL and TEAM_ID are required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := postgres.Connect(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer client.Close()
	if err := postgres.CreateTeamIfMissing(ctx, client, shared.TeamID(teamID), "Development team"); err != nil {
		return err
	}
	fmt.Println("team", teamID, "is ready")
	return nil
}
