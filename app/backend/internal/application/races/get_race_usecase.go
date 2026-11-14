package races

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// GetRaceUseCase returns a race as the caller may see it (GET /races/{race_id}). The
// create and update endpoints return the race through it too.
type GetRaceUseCase struct {
	query      RaceQueryService
	visibility race.VisibilityPolicy
	dtoFactory RaceDTOFactory
	teamID     shared.TeamID
}

// NewGetRaceUseCase returns a GetRaceUseCase for the team.
func NewGetRaceUseCase(query RaceQueryService, visibility race.VisibilityPolicy, dtoFactory RaceDTOFactory, teamID shared.TeamID) *GetRaceUseCase {
	return &GetRaceUseCase{query: query, visibility: visibility, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute returns the race. A race the caller may not see is shared.ErrNotFound, the same
// as a race that does not exist (docs/design/api.md).
func (useCase *GetRaceUseCase) Execute(ctx context.Context, caller auth.MeDTO, raceID shared.RaceID) (RaceDTO, error) {
	record, err := useCase.query.FindRace(ctx, useCase.teamID, raceID, rosterMemberOf(caller))
	if err != nil {
		return RaceDTO{}, err
	}
	visible := useCase.visibility.For(viewerOf(caller, record.Answer))
	if !visible.Race {
		return RaceDTO{}, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	description, err := race.NewDescription(record.Description)
	if err != nil {
		return RaceDTO{}, fmt.Errorf("race %s has an invalid stored description: %w", raceID, err)
	}
	return useCase.dtoFactory.Create(record.ID, record.Name, record.Version, description, record.Answer,
		record.RunningOrderConfirmed, record.RunsInConfirmedOrder, caller.IsOrganizer(), visible), nil
}
