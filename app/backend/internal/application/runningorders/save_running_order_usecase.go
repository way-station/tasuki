package runningorders

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// SaveRunningOrderUseCase saves the legs of an entry's running order
// (PUT /races/{race_id}/entries/{entry_id}/running-order). Only organizers may call it; the
// middleware checks the role.
type SaveRunningOrderUseCase struct {
	unitOfWork     transaction.UnitOfWork
	races          race.Repository
	entries        participation.EntryRepository
	runningOrders  runningorder.Repository
	history        runningorder.HistoryRepository
	laps           LapRemover
	factory        runningorder.Factory
	historyFactory runningorder.HistoryFactory
	now            func() time.Time
	teamID         shared.TeamID
}

// NewSaveRunningOrderUseCase returns a SaveRunningOrderUseCase for the team.
func NewSaveRunningOrderUseCase(unitOfWork transaction.UnitOfWork, races race.Repository, entries participation.EntryRepository, runningOrders runningorder.Repository, history runningorder.HistoryRepository, laps LapRemover, factory runningorder.Factory, historyFactory runningorder.HistoryFactory, now func() time.Time, teamID shared.TeamID) *SaveRunningOrderUseCase {
	return &SaveRunningOrderUseCase{unitOfWork: unitOfWork, races: races, entries: entries, runningOrders: runningOrders, history: history, laps: laps, factory: factory, historyFactory: historyFactory, now: now, teamID: teamID}
}

// Execute saves the first draft, or replaces the legs of the stored running order, and
// keeps one history entry, in one unit of work. In a confirmed running order, runners
// whose number of legs went down lose their laps for the race. Without a version there
// must be no stored running order, and with one there must be; otherwise it is
// shared.ErrVersionConflict. A race of another team, or an entry of another race, is
// shared.ErrNotFound. The response is read with GetRunningOrderUseCase.
func (useCase *SaveRunningOrderUseCase) Execute(ctx context.Context, caller auth.MeDTO, command SaveRunningOrderCommand) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		found, entry, err := findRaceEntry(ctx, useCase.races, useCase.entries, useCase.teamID, command.raceID, command.entryID)
		if err != nil {
			return err
		}
		runningOrder, lost, err := useCase.apply(ctx, command, shapeOf(found.Description()), entry.Members())
		if err != nil {
			return err
		}
		if err := useCase.runningOrders.Save(ctx, runningOrder); err != nil {
			return err
		}
		if err := useCase.history.Append(ctx, useCase.historyFactory.Snapshot(runningOrder, command.raceID, caller.UserID(), useCase.now())); err != nil {
			return err
		}
		if len(lost) == 0 {
			return nil
		}
		return useCase.laps.RemoveLaps(ctx, command.raceID, lost)
	})
}

// apply returns the new first draft, or the stored running order with the new legs and
// the runners who lose their laps.
func (useCase *SaveRunningOrderUseCase) apply(ctx context.Context, command SaveRunningOrderCommand, shape runningorder.Shape, entryMembers []shared.MemberID) (*runningorder.RunningOrder, []shared.MemberID, error) {
	stored, err := useCase.runningOrders.FindByEntry(ctx, command.entryID)
	if errors.Is(err, shared.ErrNotFound) {
		if command.version != nil {
			return nil, nil, fmt.Errorf("running order of entry %s was not saved yet: %w", command.entryID, shared.ErrVersionConflict)
		}
		created, err := useCase.factory.New(command.entryID, command.assignments, shape, entryMembers)
		return created, nil, err
	}
	if err != nil {
		return nil, nil, err
	}
	if command.version == nil {
		return nil, nil, fmt.Errorf("running order of entry %s was already saved: %w", command.entryID, shared.ErrVersionConflict)
	}
	lost, err := stored.Replace(command.assignments, shape, entryMembers, *command.version)
	return stored, lost, err
}
