package runningorders

import (
	"context"
	"errors"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// AdjustRunningOrdersUseCase fits the running orders of a race to its new legs. It
// implements races.RunningOrderAdjuster (#87), which the race edit calls in its unit of
// work ("Running order: legs when the race changes" in docs/design/api.md).
type AdjustRunningOrdersUseCase struct {
	entries       participation.EntryRepository
	runningOrders runningorder.Repository
	laps          LapRemover
}

// NewAdjustRunningOrdersUseCase returns an AdjustRunningOrdersUseCase.
func NewAdjustRunningOrdersUseCase(entries participation.EntryRepository, runningOrders runningorder.Repository, laps LapRemover) *AdjustRunningOrdersUseCase {
	return &AdjustRunningOrdersUseCase{entries: entries, runningOrders: runningOrders, laps: laps}
}

// AdjustRunningOrders adjusts the running order of every entry of the race. Runners who
// lose legs of a confirmed running order lose their laps for the race. No history entry is
// kept, because the backend made the change (docs/design/api.md).
func (useCase *AdjustRunningOrdersUseCase) AdjustRunningOrders(ctx context.Context, raceID shared.RaceID, description race.Description) error {
	return forEachRunningOrder(ctx, useCase.entries, useCase.runningOrders, raceID, func(runningOrder *runningorder.RunningOrder) error {
		lost := runningOrder.AdjustTo(shapeOf(description))
		if err := saveIfChanged(ctx, useCase.runningOrders, runningOrder); err != nil {
			return err
		}
		if len(lost) == 0 {
			return nil
		}
		return useCase.laps.RemoveLaps(ctx, raceID, lost)
	})
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
