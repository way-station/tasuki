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
func (lap Lap) ID() shared.LapID { return lap.id }

// Time returns the lap time per km.
func (lap Lap) Time() shared.Pace { return lap.time }

// SubmittedAt returns when the lap was submitted.
func (lap Lap) SubmittedAt() time.Time { return lap.submittedAt }

// Factory assembles Lap values.
type Factory struct {
	ids shared.IDGenerator
}

// NewFactory returns a Factory that takes new IDs from ids.
func NewFactory(ids shared.IDGenerator) Factory {
	return Factory{ids: ids}
}

// New returns a lap submitted at now.
func (factory Factory) New(lapTime shared.Pace, now time.Time) Lap {
	return Lap{id: shared.LapID(factory.ids.NewID()), time: lapTime, submittedAt: now}
}

// Reconstruct returns a lap read from storage.
func (factory Factory) Reconstruct(id shared.LapID, lapTime shared.Pace, submittedAt time.Time) Lap {
	return Lap{id: id, time: lapTime, submittedAt: submittedAt}
}

// Record is a member's laps for one race, in submission order. It is created by RecordFactory.
type Record struct {
	raceID   shared.RaceID
	memberID shared.MemberID
	laps     []Lap
}

// RaceID returns the race.
func (record *Record) RaceID() shared.RaceID { return record.raceID }

// MemberID returns the runner.
func (record *Record) MemberID() shared.MemberID { return record.memberID }

// Laps returns the laps in submission order.
func (record *Record) Laps() []Lap { return slices.Clone(record.laps) }

// Times returns the lap times in submission order.
func (record *Record) Times() []shared.Pace {
	times := make([]shared.Pace, len(record.laps))
	for i, lap := range record.laps {
		times[i] = lap.time
	}
	return times
}

// Add appends newly submitted laps.
func (record *Record) Add(laps ...Lap) {
	record.laps = append(record.laps, laps...)
}

// Edit replaces the time of a lap. It returns an error wrapping shared.ErrNotFound when the
// record has no such lap.
func (record *Record) Edit(id shared.LapID, lapTime shared.Pace) error {
	position := record.index(id)
	if position < 0 {
		return shared.ErrNotFound
	}
	record.laps[position].time = lapTime
	return nil
}

// Delete removes a lap. It returns an error wrapping shared.ErrNotFound when the record has
// no such lap.
func (record *Record) Delete(id shared.LapID) error {
	position := record.index(id)
	if position < 0 {
		return shared.ErrNotFound
	}
	record.laps = slices.Delete(record.laps, position, position+1)
	return nil
}

func (record *Record) index(id shared.LapID) int {
	return slices.IndexFunc(record.laps, func(lap Lap) bool { return lap.id == id })
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
	Save(ctx context.Context, record *Record) error
	// Delete removes the member's laps for the race, when they are removed from a confirmed
	// running order.
	Delete(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error
}
