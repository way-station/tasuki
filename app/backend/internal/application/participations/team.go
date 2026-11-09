package participations

import (
	"context"
	"fmt"

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
