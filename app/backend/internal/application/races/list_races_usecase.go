package races

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ListRacesUseCase lists the races the caller may see (GET /races).
type ListRacesUseCase struct {
	query      RaceQueryService
	visibility race.VisibilityPolicy
	dtoFactory RaceDTOFactory
	teamID     shared.TeamID
}

// NewListRacesUseCase returns a ListRacesUseCase for the team.
func NewListRacesUseCase(query RaceQueryService, visibility race.VisibilityPolicy, dtoFactory RaceDTOFactory, teamID shared.TeamID) *ListRacesUseCase {
	return &ListRacesUseCase{query: query, visibility: visibility, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute returns the races, newest registered first. Organizers see every race; members
// see the races they are invited to, except those they answered "not attending"
// (ADR-0006).
func (useCase *ListRacesUseCase) Execute(ctx context.Context, caller auth.MeDTO) ([]RaceSummaryDTO, error) {
	records, err := useCase.query.ListRaces(ctx, useCase.teamID, rosterMemberOf(caller))
	if err != nil {
		return nil, err
	}
	summaries := make([]RaceSummaryDTO, 0, len(records))
	for _, record := range records {
		visible := useCase.visibility.For(viewerOf(caller, record.Answer))
		if !visible.Race {
			continue
		}
		summaries = append(summaries, useCase.dtoFactory.CreateSummary(record.ID, record.Name, record.RaceDate, record.Answer, record.RunningOrderConfirmed, visible))
	}
	return summaries, nil
}
