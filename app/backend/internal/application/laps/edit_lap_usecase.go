package laps

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// EditLapUseCase sets the time of one of the caller's laps
// (PATCH /races/{race_id}/my-laps/{lap_id}).
type EditLapUseCase struct {
	unitOfWork   transaction.UnitOfWork
	query        LapQueryService
	records      lap.Repository
	outbox       outbox.Outbox
	policy       lap.SubmissionPolicy
	entryFactory outbox.EntryFactory
	teamID       shared.TeamID
}

// NewEditLapUseCase returns an EditLapUseCase for the team.
func NewEditLapUseCase(unitOfWork transaction.UnitOfWork, query LapQueryService, records lap.Repository, laps outbox.Outbox, policy lap.SubmissionPolicy, entryFactory outbox.EntryFactory, teamID shared.TeamID) *EditLapUseCase {
	return &EditLapUseCase{unitOfWork: unitOfWork, query: query, records: records, outbox: laps, policy: policy, entryFactory: entryFactory, teamID: teamID}
}

// Execute adds the outbox entry that sets the time of the lap and recalculates the caller's
// average (docs/design/api.md). A caller who is not a runner in a confirmed running order of
// the race is lap.ErrNotRunner, and a lap the caller does not have in the race is
// shared.ErrNotFound. The response is read with GetMyLapsUseCase.
func (useCase *EditLapUseCase) Execute(ctx context.Context, caller auth.MeDTO, command EditLapCommand) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		memberID, err := requireLap(ctx, useCase.query, useCase.records, useCase.policy, useCase.teamID, caller, command.raceID, command.lapID)
		if err != nil {
			return err
		}
		entry, err := useCase.entryFactory.New(
			EditLapStep{RaceID: command.raceID, MemberID: memberID, LapID: command.lapID, Seconds: command.lapTime.Seconds()},
			RecalculateAverageStep{MemberID: memberID},
		)
		if err != nil {
			return err
		}
		return useCase.outbox.Add(ctx, entry)
	})
}
