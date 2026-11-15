package runningorders

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// findRaceEntry returns the race and the entry. It returns an error wrapping
// shared.ErrNotFound unless the race belongs to the team and the entry to the race.
func findRaceEntry(ctx context.Context, races race.Repository, entries participation.EntryRepository, teamID shared.TeamID, raceID shared.RaceID, entryID shared.EntryID) (*race.Race, *participation.Entry, error) {
	found, err := races.FindByID(ctx, raceID)
	if err != nil {
		return nil, nil, err
	}
	if found.TeamID() != teamID {
		return nil, nil, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	entry, err := entries.FindByID(ctx, entryID)
	if err != nil {
		return nil, nil, err
	}
	if entry.RaceID() != raceID {
		return nil, nil, fmt.Errorf("entry %s of race %s: %w", entryID, raceID, shared.ErrNotFound)
	}
	return found, entry, nil
}

// shapeOf returns what the legs of a running order depend on in the race.
func shapeOf(description race.Description) runningorder.Shape {
	return runningorder.Shape{TimeBased: description.Type() == race.Time, LegCount: description.LegCount()}
}
