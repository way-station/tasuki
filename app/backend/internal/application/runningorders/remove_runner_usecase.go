package runningorders

import (
	"context"

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
