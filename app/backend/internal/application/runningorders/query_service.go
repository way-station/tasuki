package runningorders

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RunningOrderQueryService reads the running orders of a race with what the screens show
// and the predictions need. It is implemented in the infrastructure layer.
type RunningOrderQueryService interface {
	// ListRunningOrders returns the race and the saved running orders of its entries,
	// oldest entry first. rosterMemberID is whose answer to read; with nil, the answer is
	// race.NotInvited. It returns an error wrapping shared.ErrNotFound when the team has no
	// such race.
	ListRunningOrders(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID, rosterMemberID *shared.MemberID) (RaceRunningOrdersRecord, error)
}

// RaceRunningOrdersRecord is a race with its running orders as ListRunningOrders reads it.
type RaceRunningOrdersRecord struct {
	Description   race.DescriptionParams
	Answer        race.Answer
	RunningOrders []RunningOrderRecord
}

// RunningOrderRecord is the running order of one entry.
type RunningOrderRecord struct {
	EntryID   shared.EntryID
	EntryName string
	Status    runningorder.Status
	Version   int
	Legs      []LegRecord
}

// LegRecord is a leg with its runner, or nil for an empty leg.
type LegRecord struct {
	Number int
	Runner *RunnerRecord
}

// RunnerRecord is a runner with their average time, which the prediction uses.
type RunnerRecord struct {
	MemberID       shared.MemberID
	Name           string
	AverageSeconds int // per km
}
