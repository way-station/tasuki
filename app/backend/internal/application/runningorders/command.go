// Package runningorders holds the use cases of running orders and their predictions
// (docs/design/api.md, ADR-0011).
package runningorders

import (
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LegInput is a leg as the request sends it: a leg number and a runner, or nil for empty.
type LegInput struct {
	Number int
	Runner *shared.MemberID
}

// SaveRunningOrderCommand is the legs of an entry's running order as an organizer saves
// them, and the version they read (nil for the first draft). It is created by
// RunningOrderCommandFactory.
type SaveRunningOrderCommand struct {
	raceID      shared.RaceID
	entryID     shared.EntryID
	version     *int
	assignments []runningorder.Assignment
}

// ConfirmRunningOrderCommand is the running order to confirm and the version the organizer
// read. It is created by RunningOrderCommandFactory.
type ConfirmRunningOrderCommand struct {
	raceID  shared.RaceID
	entryID shared.EntryID
	version int
}

// RunningOrderCommandFactory assembles the commands of this package. The rules of the legs
// are left to the domain.
type RunningOrderCommandFactory struct{}

// Save returns the command to save the legs of the entry's running order.
func (RunningOrderCommandFactory) Save(raceID shared.RaceID, entryID shared.EntryID, version *int, legs []LegInput) SaveRunningOrderCommand {
	assignments := make([]runningorder.Assignment, 0, len(legs))
	for _, leg := range legs {
		assignments = append(assignments, runningorder.Assignment{Number: leg.Number, Runner: leg.Runner})
	}
	return SaveRunningOrderCommand{raceID: raceID, entryID: entryID, version: version, assignments: assignments}
}

// Confirm returns the command to confirm the entry's running order.
func (RunningOrderCommandFactory) Confirm(raceID shared.RaceID, entryID shared.EntryID, version int) ConfirmRunningOrderCommand {
	return ConfirmRunningOrderCommand{raceID: raceID, entryID: entryID, version: version}
}
