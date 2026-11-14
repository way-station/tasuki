package runningorders

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type saveFixture struct {
	unitOfWork *fakeUnitOfWork
	races      *fakeRaceRepository
	entries    *fakeEntryRepository
	orders     *fakeRunningOrderRepository
	history    *fakeHistoryRepository
	laps       *fakeLapRemover
}

func newSaveFixture(t *testing.T, orders *fakeRunningOrderRepository) *saveFixture {
	t.Helper()
	return &saveFixture{
		unitOfWork: &fakeUnitOfWork{},
		races:      &fakeRaceRepository{stored: teamRace(t, teamID, distanceParams(t))},
		entries:    entriesOf([]shared.MemberID{"m1", "m2"}),
		orders:     orders,
		history:    &fakeHistoryRepository{},
		laps:       &fakeLapRemover{},
	}
}

func (fixture *saveFixture) save(version *int, legs ...LegInput) error {
	command := RunningOrderCommandFactory{}.Save("r1", "e1", version, legs)
	useCase := NewSaveRunningOrderUseCase(fixture.unitOfWork, fixture.races, fixture.entries, fixture.orders, fixture.history, fixture.laps,
		runningorder.NewFactory(fixedIDs{}), runningorder.HistoryFactory{}, clock, teamID)
	return useCase.Execute(context.Background(), organizer(), command)
}

func version(value int) *int { return &value }

func TestSaveRunningOrderUseCase(t *testing.T) {
	t.Run("saves the first draft and keeps a history entry", func(t *testing.T) {
		fixture := newSaveFixture(t, ordersOf())
		if err := fixture.save(nil, LegInput{Number: 1, Runner: member("m1")}, LegInput{Number: 2}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(fixture.orders.saved) != 1 || fixture.orders.saved[0].Status() != runningorder.Draft || !fixture.orders.savedInTransaction {
			t.Fatalf("saved %+v, want a draft in the transaction", fixture.orders.saved)
		}
		if len(fixture.history.appended) != 1 || fixture.history.appended[0].ChangedBy() != "u2" || !fixture.history.appended[0].ChangedAt().Equal(testNow) || fixture.history.appended[0].RaceID() != "r1" {
			t.Errorf("history = %+v, want one entry by u2 at %v", fixture.history.appended, testNow)
		}
		if len(fixture.laps.removed) != 0 {
			t.Errorf("removed laps %+v, want none", fixture.laps.removed)
		}
	})

	t.Run("removes the laps of runners who lose legs of a confirmed order", func(t *testing.T) {
		fixture := newSaveFixture(t, ordersOf(storedOrder(runningorder.Confirmed, 3, member("m1"), member("m2"))))
		if err := fixture.save(version(3), LegInput{Number: 1, Runner: member("m2")}, LegInput{Number: 2, Runner: member("m2")}); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(fixture.orders.saved) != 1 || fixture.orders.saved[0].Status() != runningorder.Confirmed || fixture.orders.saved[0].Version() != 4 {
			t.Errorf("saved %+v, want the confirmed order at version 4", fixture.orders.saved)
		}
		if len(fixture.laps.removed) != 1 || !slices.Equal(fixture.laps.removed[0].memberIDs, []shared.MemberID{"m1"}) || fixture.laps.removed[0].raceID != "r1" || !fixture.laps.removed[0].inTransaction {
			t.Errorf("removed laps %+v, want m1's laps of r1 in the transaction", fixture.laps.removed)
		}
	})

	t.Run("rolls back when the laps cannot be removed", func(t *testing.T) {
		fixture := newSaveFixture(t, ordersOf(storedOrder(runningorder.Confirmed, 3, member("m1"), member("m2"))))
		fixture.laps.err = errDatabase
		if err := fixture.save(version(3), LegInput{Number: 1, Runner: member("m2")}, LegInput{Number: 2}); !errors.Is(err, errDatabase) || !errors.Is(fixture.unitOfWork.err, errDatabase) {
			t.Errorf("Execute() error = %v, unit of work got %v, want %v", err, fixture.unitOfWork.err, errDatabase)
		}
	})

	failures := []struct {
		name      string
		orders    *fakeRunningOrderRepository
		raceTeam  shared.TeamID // teamID unless set
		entryRace shared.RaceID // r1 unless set
		version   *int
		legs      []LegInput
		check     func(error) bool
	}{
		{name: "a version for an order not saved yet", orders: ordersOf(), version: version(0), legs: []LegInput{{Number: 1}, {Number: 2}}, check: isConflict},
		{name: "no version for a saved order", orders: ordersOf(storedOrder(runningorder.Draft, 1, nil, nil)), legs: []LegInput{{Number: 1}, {Number: 2}}, check: isConflict},
		{name: "a stale version", orders: ordersOf(storedOrder(runningorder.Draft, 2, nil, nil)), version: version(1), legs: []LegInput{{Number: 1}, {Number: 2}}, check: isConflict},
		{name: "a runner who is not in the entry", orders: ordersOf(), legs: []LegInput{{Number: 1, Runner: member("m9")}, {Number: 2}}, check: isValidation},
		{name: "the wrong number of legs", orders: ordersOf(), legs: []LegInput{{Number: 1}}, check: isValidation},
		{name: "a race of another team", orders: ordersOf(), raceTeam: "t2", legs: []LegInput{{Number: 1}, {Number: 2}}, check: isNotFound},
		{name: "an entry of another race", orders: ordersOf(), entryRace: "r2", legs: []LegInput{{Number: 1}, {Number: 2}}, check: isNotFound},
	}
	for _, testCase := range failures {
		t.Run("rejects "+testCase.name, func(t *testing.T) {
			fixture := newSaveFixture(t, testCase.orders)
			if testCase.raceTeam != "" {
				fixture.races = &fakeRaceRepository{stored: teamRace(t, testCase.raceTeam, distanceParams(t))}
			}
			if testCase.entryRace != "" {
				fixture.entries = &fakeEntryRepository{stored: []*participation.Entry{participation.NewEntryFactory(fixedIDs{}).Reconstruct("e1", testCase.entryRace, "Team A", nil)}}
			}
			if err := fixture.save(testCase.version, testCase.legs...); !testCase.check(err) {
				t.Errorf("Execute() error = %v", err)
			}
			if len(fixture.orders.saved) != 0 || len(fixture.history.appended) != 0 {
				t.Error("the running order was saved")
			}
		})
	}

	t.Run("allows a time race any number of legs", func(t *testing.T) {
		fixture := newSaveFixture(t, ordersOf())
		fixture.races = &fakeRaceRepository{stored: teamRace(t, teamID, timeParams(t))}
		if err := fixture.save(nil, LegInput{Number: 1}, LegInput{Number: 2}, LegInput{Number: 3}); err != nil {
			t.Errorf("Execute() error = %v", err)
		}
	})
}

func isConflict(err error) bool { return errors.Is(err, shared.ErrVersionConflict) }

func isNotFound(err error) bool { return errors.Is(err, shared.ErrNotFound) }
