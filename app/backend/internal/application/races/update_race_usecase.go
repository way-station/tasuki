package races

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// UpdateRaceUseCase edits a race of the team (PUT /races/{race_id}). Only organizers may
// call it; the middleware checks the role.
type UpdateRaceUseCase struct {
	unitOfWork transaction.UnitOfWork
	races      race.Repository
	adjuster   RunningOrderAdjuster
	teamID     shared.TeamID
}

// NewUpdateRaceUseCase returns an UpdateRaceUseCase for the team.
func NewUpdateRaceUseCase(unitOfWork transaction.UnitOfWork, races race.Repository, adjuster RunningOrderAdjuster, teamID shared.TeamID) *UpdateRaceUseCase {
	return &UpdateRaceUseCase{unitOfWork: unitOfWork, races: races, adjuster: adjuster, teamID: teamID}
}

// Execute saves the edit. When the race type or the number of legs changes, the running
// orders are adjusted in the same unit of work. A race of another team is
// shared.ErrNotFound. The response is read with GetRaceUseCase.
func (useCase *UpdateRaceUseCase) Execute(ctx context.Context, command UpdateRaceCommand) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		current, err := useCase.races.FindByID(ctx, command.raceID)
		if err != nil {
			return err
		}
		if current.TeamID() != useCase.teamID {
			return fmt.Errorf("race %s: %w", command.raceID, shared.ErrNotFound)
		}
		changed := legsChanged(current.Description(), command.description)
		if err := current.Update(command.name, command.description, command.version); err != nil {
			return err
		}
		if err := useCase.races.Save(ctx, current); err != nil {
			return err
		}
		if !changed {
			return nil
		}
		return useCase.adjuster.AdjustRunningOrders(ctx, current.ID(), current.Description())
	})
}

// legsChanged reports whether the running orders must be adjusted: the race type or the
// number of legs changed.
func legsChanged(before, after race.Description) bool {
	return before.Type() != after.Type() || before.LegCount() != after.LegCount()
}
