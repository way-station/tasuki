// Package lap is the domain of lap times: the laps a runner submits for a race, kept in
// MongoDB (lap_records, docs/design/database.md). The average time is in domain/roster.
package lap

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ErrNotRunner means the member is not a runner in a confirmed running order of the race.
// Only such runners can submit laps; the API returns 404.
var ErrNotRunner = errors.New("not a runner in a confirmed running order")

// Lap is one lap time per km. It is created by Factory.
type Lap struct {
	id          shared.LapID
	time        shared.Pace
	submittedAt time.Time
}

// ID returns the lap's ID.
func (l Lap) ID() shared.LapID { return l.id }

// Time returns the lap time per km.
func (l Lap) Time() shared.Pace { return l.time }

// SubmittedAt returns when the lap was submitted.
func (l Lap) SubmittedAt() time.Time { return l.submittedAt }

// Factory assembles Lap values.
type Factory struct {
	ids shared.IDGenerator
}

// NewFactory returns a Factory that takes new IDs from ids.
func NewFactory(ids shared.IDGenerator) Factory {
	return Factory{ids: ids}
}

// New returns a lap submitted at now.
func (f Factory) New(t shared.Pace, now time.Time) Lap {
	return Lap{id: shared.LapID(f.ids.NewID()), time: t, submittedAt: now}
}

// Reconstruct returns a lap read from storage.
func (f Factory) Reconstruct(id shared.LapID, t shared.Pace, submittedAt time.Time) Lap {
	return Lap{id: id, time: t, submittedAt: submittedAt}
}

// Record is a member's laps for one race, in submission order. It is created by RecordFactory.
type Record struct {
	raceID   shared.RaceID
	memberID shared.MemberID
	laps     []Lap
}

// RaceID returns the race.
func (r *Record) RaceID() shared.RaceID { return r.raceID }

// MemberID returns the runner.
func (r *Record) MemberID() shared.MemberID { return r.memberID }

// Laps returns the laps in submission order.
func (r *Record) Laps() []Lap { return slices.Clone(r.laps) }

// Times returns the lap times in submission order.
func (r *Record) Times() []shared.Pace {
	ts := make([]shared.Pace, len(r.laps))
	for i, l := range r.laps {
		ts[i] = l.time
	}
	return ts
}

// Add appends newly submitted laps.
func (r *Record) Add(laps ...Lap) {
	r.laps = append(r.laps, laps...)
}

// Edit replaces the time of a lap. It returns an error wrapping shared.ErrNotFound when the
// record has no such lap.
func (r *Record) Edit(id shared.LapID, t shared.Pace) error {
	i := r.index(id)
	if i < 0 {
		return shared.ErrNotFound
	}
	r.laps[i].time = t
	return nil
}

// Delete removes a lap. It returns an error wrapping shared.ErrNotFound when the record has
// no such lap.
func (r *Record) Delete(id shared.LapID) error {
	i := r.index(id)
	if i < 0 {
		return shared.ErrNotFound
	}
	r.laps = slices.Delete(r.laps, i, i+1)
	return nil
}

func (r *Record) index(id shared.LapID) int {
	return slices.IndexFunc(r.laps, func(l Lap) bool { return l.id == id })
}

// RecordFactory assembles Record values.
type RecordFactory struct{}

// New returns an empty record of the member's laps for the race.
func (RecordFactory) New(raceID shared.RaceID, memberID shared.MemberID) *Record {
	return &Record{raceID: raceID, memberID: memberID}
}

// Reconstruct returns a record read from storage.
func (RecordFactory) Reconstruct(raceID shared.RaceID, memberID shared.MemberID, laps []Lap) *Record {
	return &Record{raceID: raceID, memberID: memberID, laps: slices.Clone(laps)}
}

// SubmissionPolicy decides who may change laps.
type SubmissionPolicy struct{}

// CanChange returns ErrNotRunner unless the member is a runner in a confirmed running order
// of the race. Only the member themselves changes their laps; organizers only view them.
func (SubmissionPolicy) CanChange(runnerInConfirmedOrder bool) error {
	if !runnerInConfirmedOrder {
		return ErrNotRunner
	}
	return nil
}

// VisibilityPolicy decides who may see a member's laps and average time (ADR-0006).
type VisibilityPolicy struct{}

// CanSee reports whether the viewer may see the owner's laps: the owner and organizers.
func (VisibilityPolicy) CanSee(viewer shared.MemberID, organizer bool, owner shared.MemberID) bool {
	return organizer || (viewer != "" && viewer == owner)
}

// Repository stores lap records in MongoDB. Errors for missing records wrap shared.ErrNotFound.
type Repository interface {
	Find(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) (*Record, error)
	// ListByMember returns the member's records for every race, for recalculating the average.
	ListByMember(ctx context.Context, memberID shared.MemberID) ([]*Record, error)
	Save(ctx context.Context, r *Record) error
	// Delete removes the member's laps for the race, when they are removed from a confirmed
	// running order.
	Delete(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error
}
