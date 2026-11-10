package runningorders

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const teamID shared.TeamID = "t1"

var errDatabase = errors.New("database is down")

var testNow = time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC)

func clock() time.Time { return testNow }

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "o1" }

type transactionKey struct{}

func inTransaction(ctx context.Context) bool {
	marked, _ := ctx.Value(transactionKey{}).(bool)
	return marked
}

type fakeUnitOfWork struct{ err error }

func (fake *fakeUnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	fake.err = work(context.WithValue(ctx, transactionKey{}, true))
	return fake.err
}

func member(id shared.MemberID) *shared.MemberID { return &id }

// distanceParams is a distance race of two legs, 5 km and 3.2 km.
func distanceParams(t *testing.T) race.DescriptionParams {
	t.Helper()
	raceDate, err := shared.ParseDate("2026-12-06")
	if err != nil {
		t.Fatal(err)
	}
	return race.DescriptionParams{Type: race.Distance, RaceDate: raceDate, Deadline: testNow, LegDistancesKm: []float64{5, 3.2}}
}

// timeParams is a time race of 180 minutes with 2 km laps.
func timeParams(t *testing.T) race.DescriptionParams {
	t.Helper()
	params := distanceParams(t)
	params.Type, params.LegDistancesKm, params.TimeLimitMinutes, params.LapDistanceKm = race.Time, nil, 180, 2
	return params
}

func description(t *testing.T, params race.DescriptionParams) race.Description {
	t.Helper()
	described, err := race.NewDescription(params)
	if err != nil {
		t.Fatal(err)
	}
	return described
}

// teamRace returns race r1 of the team with the description.
func teamRace(t *testing.T, team shared.TeamID, params race.DescriptionParams) *race.Race {
	t.Helper()
	return race.NewFactory(fixedIDs{}).Reconstruct("r1", team, "Ekiden", description(t, params), 1)
}

type fakeRaceRepository struct{ stored *race.Race }

func (fake *fakeRaceRepository) FindByID(_ context.Context, id shared.RaceID) (*race.Race, error) {
	if fake.stored == nil || fake.stored.ID() != id {
		return nil, fmt.Errorf("race %s: %w", id, shared.ErrNotFound)
	}
	return fake.stored, nil
}

func (fake *fakeRaceRepository) Save(context.Context, *race.Race) error {
	return errors.New("races are not saved by this package")
}

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

type fakeHistoryRepository struct{ appended []*runningorder.HistoryEntry }

func (fake *fakeHistoryRepository) Append(_ context.Context, historyEntry *runningorder.HistoryEntry) error {
	fake.appended = append(fake.appended, historyEntry)
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

// linkedMember returns a member linked to roster member m1.
func linkedMember() auth.MeDTO {
	factory := auth.MeDTOFactory{}
	return factory.Create("u1", "Taro", false, identity.CallerApproved, factory.CreateRosterMember("m1", "Taro"))
}

// organizer returns an organizer without a link.
func organizer() auth.MeDTO {
	return auth.MeDTOFactory{}.Create("u2", "Hanako", true, identity.CallerNone, nil)
}

func isValidation(err error) bool {
	var validationErr *shared.ValidationError
	return errors.As(err, &validationErr)
}
