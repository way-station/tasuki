package postgres

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// LapReader reads what the lap query service needs from PostgreSQL: the runners of the
// confirmed running orders (ADR-0011). The average times are read by ProfileQueryService.
type LapReader struct {
	queries *sqlcgen.Queries
}

// NewLapReader returns a LapReader that uses client.
func NewLapReader(client *Client) *LapReader {
	return &LapReader{queries: sqlcgen.New(client.Pool())}
}

// Runner is a runner of a confirmed running order.
type Runner struct {
	MemberID shared.MemberID
	Name     string
}

// IsConfirmedRunner reports whether the roster member runs a leg of a confirmed running
// order of the race. It is false when the team has no such race.
func (reader *LapReader) IsConfirmedRunner(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID, memberID shared.MemberID) (bool, error) {
	race, err := parseID(raceID)
	if err != nil {
		return false, err
	}
	team, err := parseID(teamID)
	if err != nil {
		return false, err
	}
	member, err := parseID(memberID)
	if err != nil {
		return false, err
	}
	return queriesFor(ctx, reader.queries).IsConfirmedRunner(ctx, sqlcgen.IsConfirmedRunnerParams{ID: race, TeamID: team, RosterMemberID: &member})
}

// ConfirmedRunners returns every runner of the confirmed running orders of the race, by
// name. It returns an error wrapping shared.ErrNotFound when the team has no such race.
func (reader *LapReader) ConfirmedRunners(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]Runner, error) {
	race, err := parseID(raceID)
	if err != nil {
		return nil, err
	}
	team, err := parseID(teamID)
	if err != nil {
		return nil, err
	}
	queries := queriesFor(ctx, reader.queries)
	found, err := queries.RaceOfTeamExists(ctx, sqlcgen.RaceOfTeamExistsParams{ID: race, TeamID: team})
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	rows, err := queries.ListConfirmedRunners(ctx, race)
	if err != nil {
		return nil, err
	}
	runners := make([]Runner, 0, len(rows))
	for _, row := range rows {
		runners = append(runners, Runner{MemberID: shared.MemberID(row.ID.String()), Name: row.Name})
	}
	return runners, nil
}
