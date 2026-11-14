package runningorders

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ConfirmRunningOrderUseCase confirms an entry's running order
// (POST /races/{race_id}/entries/{entry_id}/running-order/confirm). Only organizers may call
// it; the middleware checks the role.
type ConfirmRunningOrderUseCase struct {
	unitOfWork     transaction.UnitOfWork
	races          race.Repository
	entries        participation.EntryRepository
	runningOrders  runningorder.Repository
	history        runningorder.HistoryRepository
	historyFactory runningorder.HistoryFactory
	now            func() time.Time
	teamID         shared.TeamID
}

// NewConfirmRunningOrderUseCase returns a ConfirmRunningOrderUseCase for the team.
func NewConfirmRunningOrderUseCase(unitOfWork transaction.UnitOfWork, races race.Repository, entries participation.EntryRepository, runningOrders runningorder.Repository, history runningorder.HistoryRepository, historyFactory runningorder.HistoryFactory, now func() time.Time, teamID shared.TeamID) *ConfirmRunningOrderUseCase {
	return &ConfirmRunningOrderUseCase{unitOfWork: unitOfWork, races: races, entries: entries, runningOrders: runningOrders, history: history, historyFactory: historyFactory, now: now, teamID: teamID}
}

// Execute confirms the running order and keeps one history entry, in one unit of work. An
// empty leg is a shared.ValidationError, and a confirmed running order a shared.StateError.
// A race of another team, an entry of another race, or an entry without a running order is
// shared.ErrNotFound. The response is read with GetRunningOrderUseCase.
func (useCase *ConfirmRunningOrderUseCase) Execute(ctx context.Context, caller auth.MeDTO, command ConfirmRunningOrderCommand) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		if _, _, err := findRaceEntry(ctx, useCase.races, useCase.entries, useCase.teamID, command.raceID, command.entryID); err != nil {
			return err
		}
		runningOrder, err := useCase.runningOrders.FindByEntry(ctx, command.entryID)
		if err != nil {
			return err
		}
		if err := runningOrder.Confirm(command.version); err != nil {
			return err
		}
		if err := useCase.runningOrders.Save(ctx, runningOrder); err != nil {
			return err
		}
		return useCase.history.Append(ctx, useCase.historyFactory.Snapshot(runningOrder, command.raceID, caller.UserID(), useCase.now()))
	})
}
