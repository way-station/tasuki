package runningorders

import (
	"context"
	"errors"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RemoveRunnerUseCase takes a roster member off the running orders of a race after they
// left its entry. It implements participations.RunnerRemover (#91), which is called in the
// unit of work that takes them out of the entry.
type RemoveRunnerUseCase struct {
	entries       participation.EntryRepository
	runningOrders runningorder.Repository
	laps          LapRemover
}

// NewRemoveRunnerUseCase returns a RemoveRunnerUseCase.
func NewRemoveRunnerUseCase(entries participation.EntryRepository, runningOrders runningorder.Repository, laps LapRemover) *RemoveRunnerUseCase {
	return &RemoveRunnerUseCase{entries: entries, runningOrders: runningOrders, laps: laps}
}

// RemoveRunner empties every leg the member runs. When the race has a confirmed running
// order, the member's laps for the race are deleted. That does not depend on whether a leg
// changed here: leaving the entry may already have emptied the legs through the foreign
// key (docs/design/database.md), and a member who never ran has no laps to delete. No
// history entry is kept, because the backend made the change (docs/design/api.md).
func (useCase *RemoveRunnerUseCase) RemoveRunner(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error {
	confirmed := false
	err := forEachRunningOrder(ctx, useCase.entries, useCase.runningOrders, raceID, func(runningOrder *runningorder.RunningOrder) error {
		runningOrder.RemoveRunner(memberID)
		if runningOrder.Status() == runningorder.Confirmed {
			confirmed = true
		}
		return saveIfChanged(ctx, useCase.runningOrders, runningOrder)
	})
	if err != nil || !confirmed {
		return err
	}
	return useCase.laps.RemoveLaps(ctx, raceID, []shared.MemberID{memberID})
}

// forEachRunningOrder calls change with the saved running order of each entry of the race.
func forEachRunningOrder(ctx context.Context, entries participation.EntryRepository, runningOrders runningorder.Repository, raceID shared.RaceID, change func(*runningorder.RunningOrder) error) error {
	raceEntries, err := entries.ListByRace(ctx, raceID)
	if err != nil {
		return err
	}
	for _, entry := range raceEntries {
		runningOrder, err := runningOrders.FindByEntry(ctx, entry.ID())
		if errors.Is(err, shared.ErrNotFound) {
			continue
		}
		if err != nil {
			return err
		}
		if err := change(runningOrder); err != nil {
			return err
		}
	}
	return nil
}

// saveIfChanged saves the running order when its legs changed since it was read.
func saveIfChanged(ctx context.Context, runningOrders runningorder.Repository, runningOrder *runningorder.RunningOrder) error {
	if runningOrder.Version() == runningOrder.LoadedVersion() {
		return nil
	}
	return runningOrders.Save(ctx, runningOrder)
}
