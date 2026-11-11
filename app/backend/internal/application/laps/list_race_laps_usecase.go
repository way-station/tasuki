package laps

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ListRaceLapsUseCase lists the laps of every runner in a race (GET /races/{race_id}/laps).
// Only organizers may call it; the middleware checks the role.
type ListRaceLapsUseCase struct {
	query      LapQueryService
	visibility lap.VisibilityPolicy
	dtoFactory LapDTOFactory
	teamID     shared.TeamID
}

// NewListRaceLapsUseCase returns a ListRaceLapsUseCase for the team.
func NewListRaceLapsUseCase(query LapQueryService, visibility lap.VisibilityPolicy, dtoFactory LapDTOFactory, teamID shared.TeamID) *ListRaceLapsUseCase {
	return &ListRaceLapsUseCase{query: query, visibility: visibility, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute returns the runners of the confirmed running orders by name, with the laps the
// caller may see (ADR-0006). A race of another team is shared.ErrNotFound.
func (useCase *ListRaceLapsUseCase) Execute(ctx context.Context, caller auth.MeDTO, raceID shared.RaceID) ([]RunnerLapsDTO, error) {
	records, err := useCase.query.ListRaceLaps(ctx, useCase.teamID, raceID)
	if err != nil {
		return nil, err
	}
	viewer := viewerOf(caller)
	runners := make([]RunnerLapsDTO, 0, len(records))
	for _, record := range records {
		if !useCase.visibility.CanSee(viewer, caller.IsOrganizer(), record.MemberID) {
			continue
		}
		runners = append(runners, useCase.dtoFactory.CreateRunnerLaps(record))
	}
	return runners, nil
}

// viewerOf returns the roster member the caller acts as, or "" when they are not linked.
func viewerOf(caller auth.MeDTO) shared.MemberID {
	rosterMember, ok := caller.RosterMember()
	if !ok {
		return ""
	}
	return rosterMember.ID()
}
