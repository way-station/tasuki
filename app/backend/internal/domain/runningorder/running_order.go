// Package runningorder is the domain of running orders: which roster member runs which leg
// of an entry, the draft and confirmed states, and their history (docs/design/database.md).
package runningorder

import (
	"slices"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Status is the state of a running order.
type Status string

// Running order states.
const (
	Draft     Status = "draft"
	Confirmed Status = "confirmed"
)

// Shape is what a running order's legs depend on in the race.
type Shape struct {
	TimeBased bool
	LegCount  int // distance races: the number of leg distances
}

// Leg is one leg of a running order. A leg without a runner is shown as "not decided".
type Leg struct {
	number int
	runner *shared.MemberID
}

// Number returns the leg number, from 1.
func (leg Leg) Number() int { return leg.number }

// Runner returns the runner, and false for an empty leg.
func (leg Leg) Runner() (shared.MemberID, bool) {
	if leg.runner == nil {
		return "", false
	}
	return *leg.runner, true
}

// Assignment is a leg as an organizer sends it: a leg number and a runner, or nil for empty.
type Assignment struct {
	Number int
	Runner *shared.MemberID
}

// RunningOrder is the running order of one entry. It is created by Factory.
type RunningOrder struct {
	id      shared.RunningOrderID
	entryID shared.EntryID
	status  Status
	version int
	loaded  int // the version it was read with; 0 for a new running order
	legs    []Leg
}

// ID returns the running order's ID.
func (runningOrder *RunningOrder) ID() shared.RunningOrderID { return runningOrder.id }

// EntryID returns the entry.
func (runningOrder *RunningOrder) EntryID() shared.EntryID { return runningOrder.entryID }

// Status returns draft or confirmed.
func (runningOrder *RunningOrder) Status() Status { return runningOrder.status }

// Version returns the version used to detect concurrent edits.
func (runningOrder *RunningOrder) Version() int { return runningOrder.version }

// LoadedVersion returns the version the running order was read with. Repositories store it
// only if the stored version is still this one.
func (runningOrder *RunningOrder) LoadedVersion() int { return runningOrder.loaded }

// Legs returns the legs in order.
func (runningOrder *RunningOrder) Legs() []Leg { return slices.Clone(runningOrder.legs) }

// HasEmptyLegs reports whether some leg has no runner.
func (runningOrder *RunningOrder) HasEmptyLegs() bool {
	return slices.ContainsFunc(runningOrder.legs, func(leg Leg) bool { return leg.runner == nil })
}

// Replace sets all the legs, as an organizer saves them. A confirmed running order stays
// confirmed. It returns the runners who lose their laps for the race: in a confirmed running
// order, the runners whose number of legs went down. Moving a runner to another leg keeps
// their laps.
func (runningOrder *RunningOrder) Replace(assignments []Assignment, shape Shape, entryMembers []shared.MemberID, version int) ([]shared.MemberID, error) {
	if version != runningOrder.version {
		return nil, shared.ErrVersionConflict
	}
	legs, err := buildLegs(assignments, shape, entryMembers)
	if err != nil {
		return nil, err
	}
	return runningOrder.setLegs(legs), nil
}

// Confirm confirms the running order. Every leg needs a runner.
func (runningOrder *RunningOrder) Confirm(version int) error {
	if version != runningOrder.version {
		return shared.ErrVersionConflict
	}
	if runningOrder.status == Confirmed {
		return shared.NewStateError("the running order is already confirmed")
	}
	if runningOrder.HasEmptyLegs() {
		return shared.NewValidationError("legs", "every leg needs a runner")
	}
	runningOrder.status = Confirmed
	runningOrder.version++
	return nil
}

// AdjustTo changes the legs after the race type or the number of legs changed. A distance
// race drops the last legs or adds empty legs at the end; a time race keeps the legs.
// It returns the runners who lose their laps, as Replace does.
func (runningOrder *RunningOrder) AdjustTo(shape Shape) []shared.MemberID {
	if shape.TimeBased || len(runningOrder.legs) == shape.LegCount {
		return nil
	}
	legs := slices.Clone(runningOrder.legs)
	if len(legs) > shape.LegCount {
		legs = legs[:shape.LegCount]
	}
	for number := len(legs) + 1; number <= shape.LegCount; number++ {
		legs = append(legs, Leg{number: number})
	}
	return runningOrder.setLegs(legs)
}

// RemoveRunner empties every leg the member runs, when they leave the entry. It returns
// the runners who lose their laps, as Replace does.
func (runningOrder *RunningOrder) RemoveRunner(memberID shared.MemberID) []shared.MemberID {
	legs := slices.Clone(runningOrder.legs)
	for i, leg := range legs {
		if leg.runner != nil && *leg.runner == memberID {
			legs[i].runner = nil
		}
	}
	return runningOrder.setLegs(legs)
}

// setLegs replaces the legs and returns, for a confirmed running order, the runners whose
// number of legs went down. The version goes up when anything changed.
func (runningOrder *RunningOrder) setLegs(legs []Leg) []shared.MemberID {
	if slices.EqualFunc(runningOrder.legs, legs, sameLeg) {
		return nil
	}
	var lost []shared.MemberID
	if runningOrder.status == Confirmed {
		before, after := countLegs(runningOrder.legs), countLegs(legs)
		for _, leg := range runningOrder.legs {
			if leg.runner == nil {
				continue
			}
			id := *leg.runner
			if after[id] < before[id] && !slices.Contains(lost, id) {
				lost = append(lost, id)
			}
		}
	}
	runningOrder.legs = legs
	runningOrder.version++
	return lost
}

func sameLeg(first, second Leg) bool {
	if first.number != second.number || (first.runner == nil) != (second.runner == nil) {
		return false
	}
	return first.runner == nil || *first.runner == *second.runner
}

func countLegs(legs []Leg) map[shared.MemberID]int {
	counts := map[shared.MemberID]int{}
	for _, leg := range legs {
		if leg.runner != nil {
			counts[*leg.runner]++
		}
	}
	return counts
}

// buildLegs validates the assignments: distance races have exactly one leg per leg
// distance; time races have one or more legs; legs are numbered from 1 without gaps; runners
// are entry members.
func buildLegs(assignments []Assignment, shape Shape, entryMembers []shared.MemberID) ([]Leg, error) {
	if !shape.TimeBased && len(assignments) != shape.LegCount {
		return nil, shared.NewValidationError("legs", "a distance race needs one leg per leg distance")
	}
	if len(assignments) == 0 {
		return nil, shared.NewValidationError("legs", "needs at least one leg")
	}
	legs := make([]Leg, len(assignments))
	for i, assignment := range assignments {
		if assignment.Number != i+1 {
			return nil, shared.NewValidationError("legs", "legs must be numbered from 1 without gaps")
		}
		var runner *shared.MemberID
		if assignment.Runner != nil {
			if !slices.Contains(entryMembers, *assignment.Runner) {
				return nil, shared.NewValidationError("legs", "runners must be members of the entry")
			}
			id := *assignment.Runner
			runner = &id
		}
		legs[i] = Leg{number: assignment.Number, runner: runner}
	}
	return legs, nil
}

// Factory assembles RunningOrder values.
type Factory struct {
	ids shared.IDGenerator
}

// NewFactory returns a Factory that takes new IDs from ids.
func NewFactory(ids shared.IDGenerator) Factory {
	return Factory{ids: ids}
}

// New returns the first draft of an entry's running order.
func (factory Factory) New(entryID shared.EntryID, assignments []Assignment, shape Shape, entryMembers []shared.MemberID) (*RunningOrder, error) {
	legs, err := buildLegs(assignments, shape, entryMembers)
	if err != nil {
		return nil, err
	}
	return &RunningOrder{id: shared.RunningOrderID(factory.ids.NewID()), entryID: entryID, status: Draft, legs: legs}, nil
}

// Reconstruct returns a running order read from storage.
func (factory Factory) Reconstruct(id shared.RunningOrderID, entryID shared.EntryID, status Status, version int, legs []Assignment) *RunningOrder {
	restored := make([]Leg, len(legs))
	for i, assignment := range legs {
		var runner *shared.MemberID
		if assignment.Runner != nil {
			id := *assignment.Runner
			runner = &id
		}
		restored[i] = Leg{number: assignment.Number, runner: runner}
	}
	return &RunningOrder{id: id, entryID: entryID, status: status, version: version, loaded: version, legs: restored}
}
