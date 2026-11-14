package laps

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// DeleteLapUseCase deletes one of the caller's laps
// (DELETE /races/{race_id}/my-laps/{lap_id}).
type DeleteLapUseCase struct {
	unitOfWork   transaction.UnitOfWork
	query        LapQueryService
	records      lap.Repository
	outbox       outbox.Outbox
	policy       lap.SubmissionPolicy
	entryFactory outbox.EntryFactory
	teamID       shared.TeamID
}

// NewDeleteLapUseCase returns a DeleteLapUseCase for the team.
func NewDeleteLapUseCase(unitOfWork transaction.UnitOfWork, query LapQueryService, records lap.Repository, laps outbox.Outbox, policy lap.SubmissionPolicy, entryFactory outbox.EntryFactory, teamID shared.TeamID) *DeleteLapUseCase {
	return &DeleteLapUseCase{unitOfWork: unitOfWork, query: query, records: records, outbox: laps, policy: policy, entryFactory: entryFactory, teamID: teamID}
}

// Execute adds the outbox entry that deletes the lap and recalculates the caller's average
// (docs/design/api.md). A caller who is not a runner in a confirmed running order of the race
// is lap.ErrNotRunner, and a lap the caller does not have in the race is shared.ErrNotFound.
// The response is read with GetMyLapsUseCase.
func (useCase *DeleteLapUseCase) Execute(ctx context.Context, caller auth.MeDTO, command DeleteLapCommand) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		memberID, err := requireLap(ctx, useCase.query, useCase.records, useCase.policy, useCase.teamID, caller, command.raceID, command.lapID)
		if err != nil {
			return err
		}
		entry, err := useCase.entryFactory.New(
			DeleteLapStep{RaceID: command.raceID, MemberID: memberID, LapID: command.lapID},
			RecalculateAverageStep{MemberID: memberID},
		)
		if err != nil {
			return err
		}
		return useCase.outbox.Add(ctx, entry)
	})
}
