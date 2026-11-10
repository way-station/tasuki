package runningorders

import (
	"context"
	"errors"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "o1" }

type transactionKey struct{}

func inTransaction(ctx context.Context) bool {
	marked, _ := ctx.Value(transactionKey{}).(bool)
	return marked
}

func member(id shared.MemberID) *shared.MemberID { return &id }

type fakeEntryRepository struct{ stored []*participation.Entry }

// entriesOf returns entry e1 of r1 with the members, and the other entries.
func entriesOf(members []shared.MemberID, others ...*participation.Entry) *fakeEntryRepository {
	return &fakeEntryRepository{stored: append([]*participation.Entry{participation.NewEntryFactory(fixedIDs{}).Reconstruct("e1", "r1", "Team A", members)}, others...)}
}

func (fake *fakeEntryRepository) FindByID(_ context.Context, id shared.EntryID) (*participation.Entry, error) {
	for _, entry := range fake.stored {
		if entry.ID() == id {
			return entry, nil
		}
	}
	return nil, fmt.Errorf("entry %s: %w", id, shared.ErrNotFound)
}

func (fake *fakeEntryRepository) ListByRace(_ context.Context, raceID shared.RaceID) ([]*participation.Entry, error) {
	var entries []*participation.Entry
	for _, entry := range fake.stored {
		if entry.RaceID() == raceID {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

func (fake *fakeEntryRepository) Save(context.Context, *participation.Entry) error {
	return errors.New("entries are not saved by this package")
}

type fakeRunningOrderRepository struct {
	stored             map[shared.EntryID]*runningorder.RunningOrder
	saved              []*runningorder.RunningOrder
	savedInTransaction bool
}

// storedOrder returns the running order o1 of entry e1 with a runner per leg (nil for empty).
func storedOrder(status runningorder.Status, version int, runners ...*shared.MemberID) *runningorder.RunningOrder {
	legs := make([]runningorder.Assignment, 0, len(runners))
	for i, runner := range runners {
		legs = append(legs, runningorder.Assignment{Number: i + 1, Runner: runner})
	}
	return runningorder.NewFactory(fixedIDs{}).Reconstruct("o1", "e1", status, version, legs)
}

func ordersOf(orders ...*runningorder.RunningOrder) *fakeRunningOrderRepository {
	fake := &fakeRunningOrderRepository{stored: map[shared.EntryID]*runningorder.RunningOrder{}}
	for _, order := range orders {
		fake.stored[order.EntryID()] = order
	}
	return fake
}

func (fake *fakeRunningOrderRepository) FindByEntry(_ context.Context, entryID shared.EntryID) (*runningorder.RunningOrder, error) {
	order, ok := fake.stored[entryID]
	if !ok {
		return nil, fmt.Errorf("running order of %s: %w", entryID, shared.ErrNotFound)
	}
	return order, nil
}

func (fake *fakeRunningOrderRepository) Save(ctx context.Context, order *runningorder.RunningOrder) error {
	fake.saved = append(fake.saved, order)
	fake.savedInTransaction = inTransaction(ctx)
	return nil
}

type removedLaps struct {
	raceID        shared.RaceID
	memberIDs     []shared.MemberID
	inTransaction bool
}

type fakeLapRemover struct {
	removed []removedLaps
	err     error
}

func (fake *fakeLapRemover) RemoveLaps(ctx context.Context, raceID shared.RaceID, memberIDs []shared.MemberID) error {
	fake.removed = append(fake.removed, removedLaps{raceID: raceID, memberIDs: memberIDs, inTransaction: inTransaction(ctx)})
	return fake.err
}
