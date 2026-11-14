package participations

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// requireTeamRace returns an error wrapping shared.ErrNotFound unless the race exists and
// belongs to the team.
func requireTeamRace(ctx context.Context, races race.Repository, teamID shared.TeamID, raceID shared.RaceID) error {
	found, err := races.FindByID(ctx, raceID)
	if err != nil {
		return err
	}
	if found.TeamID() != teamID {
		return fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	return nil
}

// findRaceEntry returns the entry of the command. It returns an error wrapping
// shared.ErrNotFound unless the race belongs to the team and the entry to the race.
func findRaceEntry(ctx context.Context, races race.Repository, entries participation.EntryRepository, teamID shared.TeamID, command EntryMemberCommand) (*participation.Entry, error) {
	if err := requireTeamRace(ctx, races, teamID, command.raceID); err != nil {
		return nil, err
	}
	entry, err := entries.FindByID(ctx, command.entryID)
	if err != nil {
		return nil, err
	}
	if entry.RaceID() != command.raceID {
		return nil, fmt.Errorf("entry %s of race %s: %w", command.entryID, command.raceID, shared.ErrNotFound)
	}
	return entry, nil
}
