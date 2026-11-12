package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// RequireTeam returns an error when the team is not in teams. The server checks it at
// startup, because the team's row is created by hand (docs/design/api.md).
func RequireTeam(ctx context.Context, client *Client, teamID shared.TeamID) error {
	id, err := parseID(teamID)
	if err != nil {
		return err
	}
	_, err = sqlcgen.New(client.Pool()).GetTeam(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("team %s is not in teams: create its row first (make db-seed in development)", teamID)
	}
	return err
}

// CreateTeamIfMissing creates the team's row unless it exists. Only make db-seed uses it,
// for the development team.
func CreateTeamIfMissing(ctx context.Context, client *Client, teamID shared.TeamID, name string) error {
	id, err := parseID(teamID)
	if err != nil {
		return err
	}
	return sqlcgen.New(client.Pool()).CreateTeamIfMissing(ctx, sqlcgen.CreateTeamIfMissingParams{ID: id, Name: name})
}
