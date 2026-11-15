package races

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RunningOrderAdjuster fits the running orders of a race to its new legs, after the race
// type or the number of legs changed ("Running order: legs when the race changes" in
// docs/design/api.md). It is called in the unit of work of the race edit, and is
// implemented by the running order application layer (#95).
type RunningOrderAdjuster interface {
	AdjustRunningOrders(ctx context.Context, raceID shared.RaceID, description race.Description) error
}
