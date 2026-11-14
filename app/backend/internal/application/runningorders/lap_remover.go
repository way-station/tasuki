package runningorders

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LapRemover deletes the laps of runners for a race and recalculates their average times
// from the remaining laps, after they lost their laps (docs/design/api.md). It is called
// in the unit of work that changes the running order, and is implemented by the lap
// application layer (#99).
type LapRemover interface {
	RemoveLaps(ctx context.Context, raceID shared.RaceID, memberIDs []shared.MemberID) error
}
