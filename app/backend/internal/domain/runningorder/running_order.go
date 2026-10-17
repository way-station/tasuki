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
func (l Leg) Number() int { return l.number }

// Runner returns the runner, and false for an empty leg.
func (l Leg) Runner() (shared.MemberID, bool) {
	if l.runner == nil {
		return "", false
	}
	return *l.runner, true
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
func (o *RunningOrder) ID() shared.RunningOrderID { return o.id }

// EntryID returns the entry.
func (o *RunningOrder) EntryID() shared.EntryID { return o.entryID }

// Status returns draft or confirmed.
func (o *RunningOrder) Status() Status { return o.status }

// Version returns the version used to detect concurrent edits.
func (o *RunningOrder) Version() int { return o.version }

// LoadedVersion returns the version the running order was read with. Repositories store it
// only if the stored version is still this one.
func (o *RunningOrder) LoadedVersion() int { return o.loaded }

// Legs returns the legs in order.
func (o *RunningOrder) Legs() []Leg { return slices.Clone(o.legs) }

// HasEmptyLegs reports whether some leg has no runner.
func (o *RunningOrder) HasEmptyLegs() bool {
	return slices.ContainsFunc(o.legs, func(l Leg) bool { return l.runner == nil })
}

// Replace sets all the legs, as an organizer saves them. A confirmed running order stays
// confirmed. It returns the runners who lose their laps for the race: in a confirmed running
// order, the runners whose number of legs went down. Moving a runner to another leg keeps
// their laps.
func (o *RunningOrder) Replace(as []Assignment, shape Shape, entryMembers []shared.MemberID, version int) ([]shared.MemberID, error) {
	if version != o.version {
		return nil, shared.ErrVersionConflict
	}
	legs, err := buildLegs(as, shape, entryMembers)
	if err != nil {
		return nil, err
	}
	return o.setLegs(legs), nil
}

// Confirm confirms the running order. Every leg needs a runner.
func (o *RunningOrder) Confirm(version int) error {
	if version != o.version {
		return shared.ErrVersionConflict
	}
	if o.status == Confirmed {
		return shared.NewStateError("the running order is already confirmed")
	}
	if o.HasEmptyLegs() {
		return shared.NewValidationError("legs", "every leg needs a runner")
	}
	o.status = Confirmed
	o.version++
	return nil
}

// AdjustTo changes the legs after the race type or the number of legs changed. A distance
// race drops the last legs or adds empty legs at the end; a time race keeps the legs.
// It returns the runners who lose their laps, as Replace does.
func (o *RunningOrder) AdjustTo(shape Shape) []shared.MemberID {
	if shape.TimeBased || len(o.legs) == shape.LegCount {
		return nil
	}
	legs := slices.Clone(o.legs)
	if len(legs) > shape.LegCount {
		legs = legs[:shape.LegCount]
	}
	for n := len(legs) + 1; n <= shape.LegCount; n++ {
		legs = append(legs, Leg{number: n})
	}
	return o.setLegs(legs)
}

// RemoveRunner empties every leg the member runs, when they leave the entry. It returns
// the runners who lose their laps, as Replace does.
func (o *RunningOrder) RemoveRunner(memberID shared.MemberID) []shared.MemberID {
	legs := slices.Clone(o.legs)
	for i, l := range legs {
		if l.runner != nil && *l.runner == memberID {
			legs[i].runner = nil
		}
	}
	return o.setLegs(legs)
}

// setLegs replaces the legs and returns, for a confirmed running order, the runners whose
// number of legs went down. The version goes up when anything changed.
func (o *RunningOrder) setLegs(legs []Leg) []shared.MemberID {
	if slices.EqualFunc(o.legs, legs, sameLeg) {
		return nil
	}
	var lost []shared.MemberID
	if o.status == Confirmed {
		before, after := countLegs(o.legs), countLegs(legs)
		for _, l := range o.legs {
			if l.runner == nil {
				continue
			}
			id := *l.runner
			if after[id] < before[id] && !slices.Contains(lost, id) {
				lost = append(lost, id)
			}
		}
	}
	o.legs = legs
	o.version++
	return lost
}

func sameLeg(a, b Leg) bool {
	if a.number != b.number || (a.runner == nil) != (b.runner == nil) {
		return false
	}
	return a.runner == nil || *a.runner == *b.runner
}

func countLegs(legs []Leg) map[shared.MemberID]int {
	n := map[shared.MemberID]int{}
	for _, l := range legs {
		if l.runner != nil {
			n[*l.runner]++
		}
	}
	return n
}

// buildLegs validates the assignments: distance races have exactly one leg per leg
// distance; time races have one or more legs; legs are numbered from 1 without gaps; runners
// are entry members.
func buildLegs(as []Assignment, shape Shape, entryMembers []shared.MemberID) ([]Leg, error) {
	if !shape.TimeBased && len(as) != shape.LegCount {
		return nil, shared.NewValidationError("legs", "a distance race needs one leg per leg distance")
	}
	if len(as) == 0 {
		return nil, shared.NewValidationError("legs", "needs at least one leg")
	}
	legs := make([]Leg, len(as))
	for i, a := range as {
		if a.Number != i+1 {
			return nil, shared.NewValidationError("legs", "legs must be numbered from 1 without gaps")
		}
		var runner *shared.MemberID
		if a.Runner != nil {
			if !slices.Contains(entryMembers, *a.Runner) {
				return nil, shared.NewValidationError("legs", "runners must be members of the entry")
			}
			id := *a.Runner
			runner = &id
		}
		legs[i] = Leg{number: a.Number, runner: runner}
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
func (f Factory) New(entryID shared.EntryID, as []Assignment, shape Shape, entryMembers []shared.MemberID) (*RunningOrder, error) {
	legs, err := buildLegs(as, shape, entryMembers)
	if err != nil {
		return nil, err
	}
	return &RunningOrder{id: shared.RunningOrderID(f.ids.NewID()), entryID: entryID, status: Draft, legs: legs}, nil
}

// Reconstruct returns a running order read from storage.
func (f Factory) Reconstruct(id shared.RunningOrderID, entryID shared.EntryID, status Status, version int, legs []Assignment) *RunningOrder {
	ls := make([]Leg, len(legs))
	for i, a := range legs {
		var runner *shared.MemberID
		if a.Runner != nil {
			id := *a.Runner
			runner = &id
		}
		ls[i] = Leg{number: a.Number, runner: runner}
	}
	return &RunningOrder{id: id, entryID: entryID, status: status, version: version, loaded: version, legs: ls}
}
