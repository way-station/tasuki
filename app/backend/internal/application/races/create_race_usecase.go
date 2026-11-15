package races

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// CreateRaceUseCase registers a race of the team (POST /races). Only organizers may call
// it; the middleware checks the role.
type CreateRaceUseCase struct {
	races   race.Repository
	factory race.Factory
	teamID  shared.TeamID
}

// NewCreateRaceUseCase returns a CreateRaceUseCase for the team.
func NewCreateRaceUseCase(races race.Repository, factory race.Factory, teamID shared.TeamID) *CreateRaceUseCase {
	return &CreateRaceUseCase{races: races, factory: factory, teamID: teamID}
}

// Execute registers the race and returns its ID. The response is read with GetRaceUseCase.
func (useCase *CreateRaceUseCase) Execute(ctx context.Context, command CreateRaceCommand) (shared.RaceID, error) {
	created, err := useCase.factory.New(useCase.teamID, command.name, command.description)
	if err != nil {
		return "", err
	}
	if err := useCase.races.Save(ctx, created); err != nil {
		return "", err
	}
	return created.ID(), nil
}
