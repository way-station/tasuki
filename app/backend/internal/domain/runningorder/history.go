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
func (historyEntry *HistoryEntry) RunningOrderID() shared.RunningOrderID {
	return historyEntry.runningOrderID
}

// RaceID returns the race.
func (historyEntry *HistoryEntry) RaceID() shared.RaceID { return historyEntry.raceID }

// EntryID returns the entry.
func (historyEntry *HistoryEntry) EntryID() shared.EntryID { return historyEntry.entryID }

// ChangedBy returns the organizer's account.
func (historyEntry *HistoryEntry) ChangedBy() shared.AccountID { return historyEntry.changedBy }

// ChangedAt returns when the running order was saved.
func (historyEntry *HistoryEntry) ChangedAt() time.Time { return historyEntry.changedAt }

// Status returns the state after the save.
func (historyEntry *HistoryEntry) Status() Status { return historyEntry.status }

// Legs returns the legs after the save.
func (historyEntry *HistoryEntry) Legs() []Leg { return append([]Leg(nil), historyEntry.legs...) }

// HistoryFactory assembles HistoryEntry values.
type HistoryFactory struct{}

// Snapshot returns the history entry for the running order as it is now.
func (HistoryFactory) Snapshot(runningOrder *RunningOrder, raceID shared.RaceID, changedBy shared.AccountID, changedAt time.Time) *HistoryEntry {
	return &HistoryEntry{
		runningOrderID: runningOrder.id,
		raceID:         raceID,
		entryID:        runningOrder.entryID,
		changedBy:      changedBy,
		changedAt:      changedAt,
		status:         runningOrder.status,
		legs:           runningOrder.Legs(),
	}
}

// Repository stores running orders. Errors for missing ones wrap shared.ErrNotFound.
type Repository interface {
	FindByEntry(ctx context.Context, entryID shared.EntryID) (*RunningOrder, error)
	// Save inserts or updates the running order. An update stores it only if the stored
	// version is LoadedVersion, and returns shared.ErrVersionConflict otherwise.
	Save(ctx context.Context, runningOrder *RunningOrder) error
}

// HistoryRepository appends history entries.
type HistoryRepository interface {
	Append(ctx context.Context, historyEntry *HistoryEntry) error
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
func (VisibilityPolicy) CanSee(viewer Viewer, status Status) bool {
	if viewer.Organizer {
		return true
	}
	return status == Confirmed && viewer.Attending
}
