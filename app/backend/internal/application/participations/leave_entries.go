package participations

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// leaveEntries takes the member out of the entries of the race and, if they were in one,
// off its running orders. It is called in a unit of work.
func leaveEntries(ctx context.Context, entries participation.EntryRepository, runners RunnerRemover, raceID shared.RaceID, memberID shared.MemberID) error {
	raceEntries, err := entries.ListByRace(ctx, raceID)
	if err != nil {
		return err
	}
	left := false
	for _, entry := range raceEntries {
		if !entry.Remove(memberID) {
			continue
		}
		if err := entries.Save(ctx, entry); err != nil {
			return err
		}
		left = true
	}
	if !left {
		return nil
	}
	return runners.RemoveRunner(ctx, raceID, memberID)
}
