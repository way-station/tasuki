package participations

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RunnerRemover takes a roster member off every leg of the running orders of a race, after
// they left its entry. When a running order is confirmed, their laps for the race are
// deleted and their average time is recalculated from the remaining laps
// (docs/design/api.md). It is called in the unit of work that takes them out of the entry,
// and is implemented by the running order (#95) and lap (#99) application layers.
type RunnerRemover interface {
	RemoveRunner(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error
}
