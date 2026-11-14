package laps

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RemoveLapsUseCase deletes the laps of runners for a race and recalculates their average
// times, after they were taken off a confirmed running order. It implements
// runningorders.LapRemover (#95), which is called in the unit of work that changes the
// running order; the laps are in MongoDB, so they are deleted through the outbox
// (docs/design/api.md, ADR-0016).
type RemoveLapsUseCase struct {
	steps        outbox.Outbox
	entryFactory outbox.EntryFactory
}

// NewRemoveLapsUseCase returns a RemoveLapsUseCase.
func NewRemoveLapsUseCase(steps outbox.Outbox, entryFactory outbox.EntryFactory) *RemoveLapsUseCase {
	return &RemoveLapsUseCase{steps: steps, entryFactory: entryFactory}
}

// RemoveLaps adds one outbox entry for each member: it deletes the member's laps for the
// race, and then recalculates their average time from all of their remaining laps.
func (useCase *RemoveLapsUseCase) RemoveLaps(ctx context.Context, raceID shared.RaceID, memberIDs []shared.MemberID) error {
	for _, memberID := range memberIDs {
		entry, err := useCase.entryFactory.New(
			RemoveRaceLapsStep{RaceID: raceID, MemberID: memberID},
			RecalculateAverageStep{MemberID: memberID},
		)
		if err != nil {
			return err
		}
		if err := useCase.steps.Add(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}
