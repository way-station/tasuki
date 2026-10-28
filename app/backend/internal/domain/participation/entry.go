package participation

import (
	"context"
	"slices"
	"strings"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// MaxEntriesPerRace is the number of entries a race can have in the MVP. The tables allow more.
const MaxEntriesPerRace = 1

// Entry is a group of attending members that an organizer enters into a race (for example
// "Team A"). Entries cannot be renamed or deleted. It is created by EntryFactory.
type Entry struct {
	id      shared.EntryID
	raceID  shared.RaceID
	name    string
	members []shared.MemberID
}

// ID returns the entry's ID.
func (entry *Entry) ID() shared.EntryID { return entry.id }

// RaceID returns the race.
func (entry *Entry) RaceID() shared.RaceID { return entry.raceID }

// Name returns the entry name.
func (entry *Entry) Name() string { return entry.name }

// Members returns the entry's members.
func (entry *Entry) Members() []shared.MemberID { return slices.Clone(entry.members) }

// Has reports whether the member is in the entry.
func (entry *Entry) Has(memberID shared.MemberID) bool {
	return slices.Contains(entry.members, memberID)
}

// Add puts a member in the entry. Only members whose answer is "attending" can be added.
// Adding a member who is already in the entry does nothing.
func (entry *Entry) Add(memberID shared.MemberID, answer AttendanceStatus) error {
	if answer != Attending {
		return shared.NewValidationError("roster_member_id", `the member has not answered "attending"`)
	}
	if !entry.Has(memberID) {
		entry.members = append(entry.members, memberID)
	}
	return nil
}

// Remove takes a member out of the entry and reports whether they were in it. The caller
// then removes them from the running order.
func (entry *Entry) Remove(memberID shared.MemberID) bool {
	index := slices.Index(entry.members, memberID)
	if index < 0 {
		return false
	}
	entry.members = slices.Delete(entry.members, index, index+1)
	return true
}

// EntryFactory assembles Entry values.
type EntryFactory struct {
	ids shared.IDGenerator
}

// NewEntryFactory returns an EntryFactory that takes new IDs from ids.
func NewEntryFactory(ids shared.IDGenerator) EntryFactory {
	return EntryFactory{ids: ids}
}

// New returns a new, empty entry of the race. existing is the number of entries the race
// already has; the MVP allows MaxEntriesPerRace.
func (factory EntryFactory) New(raceID shared.RaceID, name string, existing int) (*Entry, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, shared.NewValidationError("name", "is required")
	}
	if existing >= MaxEntriesPerRace {
		return nil, shared.NewStateError("the race already has an entry")
	}
	return &Entry{id: shared.EntryID(factory.ids.NewID()), raceID: raceID, name: name}, nil
}

// Reconstruct returns an entry read from storage.
func (factory EntryFactory) Reconstruct(id shared.EntryID, raceID shared.RaceID, name string, members []shared.MemberID) *Entry {
	return &Entry{id: id, raceID: raceID, name: name, members: slices.Clone(members)}
}

// EntryRepository stores entries and their members. Errors for missing ones wrap
// shared.ErrNotFound.
type EntryRepository interface {
	FindByID(ctx context.Context, id shared.EntryID) (*Entry, error)
	ListByRace(ctx context.Context, raceID shared.RaceID) ([]*Entry, error)
	Save(ctx context.Context, entry *Entry) error
}
