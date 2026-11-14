package laps

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// GetMyLapsUseCase returns the caller's laps for a race (GET /races/{race_id}/my-laps). The
// submit, edit, and delete endpoints return the laps through it.
type GetMyLapsUseCase struct {
	query      LapQueryService
	dtoFactory LapDTOFactory
	teamID     shared.TeamID
}

// NewGetMyLapsUseCase returns a GetMyLapsUseCase for the team.
func NewGetMyLapsUseCase(query LapQueryService, dtoFactory LapDTOFactory, teamID shared.TeamID) *GetMyLapsUseCase {
	return &GetMyLapsUseCase{query: query, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute returns the laps with the caller's average time. A caller who is not a runner in
// a confirmed running order of the race is lap.ErrNotRunner.
func (useCase *GetMyLapsUseCase) Execute(ctx context.Context, caller auth.MeDTO, raceID shared.RaceID) (MyLapsDTO, error) {
	memberID, runner, err := confirmedRunner(ctx, useCase.query, useCase.teamID, caller, raceID)
	if err != nil {
		return MyLapsDTO{}, err
	}
	if !runner {
		return MyLapsDTO{}, fmt.Errorf("race %s: %w", raceID, lap.ErrNotRunner)
	}
	record, err := useCase.query.GetMyLaps(ctx, raceID, memberID)
	if err != nil {
		return MyLapsDTO{}, err
	}
	return useCase.dtoFactory.CreateMyLaps(record), nil
}
