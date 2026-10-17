package runningorder

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// HistoryEntry is a snapshot of a running order at one save, kept in MongoDB
// (running_order_history). It is created by HistoryFactory.
type HistoryEntry struct {
	runningOrderID shared.RunningOrderID
	raceID         shared.RaceID
	entryID        shared.EntryID
	changedBy      shared.AccountID
	changedAt      time.Time
	status         Status
	legs           []Leg
}

// RunningOrderID returns the running order.
func (h *HistoryEntry) RunningOrderID() shared.RunningOrderID { return h.runningOrderID }

// RaceID returns the race.
func (h *HistoryEntry) RaceID() shared.RaceID { return h.raceID }

// EntryID returns the entry.
func (h *HistoryEntry) EntryID() shared.EntryID { return h.entryID }

// ChangedBy returns the organizer's account.
func (h *HistoryEntry) ChangedBy() shared.AccountID { return h.changedBy }

// ChangedAt returns when the running order was saved.
func (h *HistoryEntry) ChangedAt() time.Time { return h.changedAt }

// Status returns the state after the save.
func (h *HistoryEntry) Status() Status { return h.status }

// Legs returns the legs after the save.
func (h *HistoryEntry) Legs() []Leg { return append([]Leg(nil), h.legs...) }

// HistoryFactory assembles HistoryEntry values.
type HistoryFactory struct{}

// Snapshot returns the history entry for the running order as it is now.
func (HistoryFactory) Snapshot(o *RunningOrder, raceID shared.RaceID, by shared.AccountID, at time.Time) *HistoryEntry {
	return &HistoryEntry{
		runningOrderID: o.id,
		raceID:         raceID,
		entryID:        o.entryID,
		changedBy:      by,
		changedAt:      at,
		status:         o.status,
		legs:           o.Legs(),
	}
}

// Repository stores running orders. Errors for missing ones wrap shared.ErrNotFound.
type Repository interface {
	FindByEntry(ctx context.Context, entryID shared.EntryID) (*RunningOrder, error)
	// Save inserts or updates the running order. An update stores it only if the stored
	// version is LoadedVersion, and returns shared.ErrVersionConflict otherwise.
	Save(ctx context.Context, o *RunningOrder) error
}

// HistoryRepository appends history entries.
type HistoryRepository interface {
	Append(ctx context.Context, h *HistoryEntry) error
}

// Viewer is who is looking at a running order.
type Viewer struct {
	Organizer bool
	Attending bool // the viewer answered "attending" for the race
}

// VisibilityPolicy decides who may see a running order and its prediction (ADR-0006).
type VisibilityPolicy struct{}

// CanSee reports whether the viewer may see a running order in the given state:
// drafts only organizers; confirmed ones organizers and members who answered "attending".
func (VisibilityPolicy) CanSee(v Viewer, s Status) bool {
	if v.Organizer {
		return true
	}
	return s == Confirmed && v.Attending
}
