package laps

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// SubmitLapsUseCase submits the caller's laps for a race (POST /races/{race_id}/my-laps).
type SubmitLapsUseCase struct {
	unitOfWork   transaction.UnitOfWork
	query        LapQueryService
	outbox       outbox.Outbox
	policy       lap.SubmissionPolicy
	lapFactory   lap.Factory
	entryFactory outbox.EntryFactory
	now          func() time.Time
	teamID       shared.TeamID
}

// NewSubmitLapsUseCase returns a SubmitLapsUseCase for the team.
func NewSubmitLapsUseCase(unitOfWork transaction.UnitOfWork, query LapQueryService, laps outbox.Outbox, policy lap.SubmissionPolicy, lapFactory lap.Factory, entryFactory outbox.EntryFactory, now func() time.Time, teamID shared.TeamID) *SubmitLapsUseCase {
	return &SubmitLapsUseCase{unitOfWork: unitOfWork, query: query, outbox: laps, policy: policy, lapFactory: lapFactory, entryFactory: entryFactory, now: now, teamID: teamID}
}

// Execute gives the laps their IDs and adds the outbox entry that adds them to the lap
// record and recalculates the caller's average (docs/design/api.md). A caller who is not a
// runner in a confirmed running order of the race is lap.ErrNotRunner. The response is read
// with GetMyLapsUseCase.
func (useCase *SubmitLapsUseCase) Execute(ctx context.Context, caller auth.MeDTO, command SubmitLapsCommand) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		memberID, runner, err := confirmedRunner(ctx, useCase.query, useCase.teamID, caller, command.raceID)
		if err != nil {
			return err
		}
		if err := useCase.policy.CanChange(runner); err != nil {
			return err
		}
		now := useCase.now()
		stepLaps := make([]StepLap, 0, len(command.times))
		for _, lapTime := range command.times {
			submitted := useCase.lapFactory.New(lapTime, now)
			stepLaps = append(stepLaps, StepLap{ID: submitted.ID(), Seconds: submitted.Time().Seconds(), SubmittedAt: submitted.SubmittedAt()})
		}
		entry, err := useCase.entryFactory.New(
			AddLapsStep{RaceID: command.raceID, MemberID: memberID, Laps: stepLaps},
			RecalculateAverageStep{MemberID: memberID},
		)
		if err != nil {
			return err
		}
		return useCase.outbox.Add(ctx, entry)
	})
}
