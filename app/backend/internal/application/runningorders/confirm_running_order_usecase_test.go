package runningorders

import (
	"context"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
)

func TestConfirmRunningOrderUseCase(t *testing.T) {
	confirm := func(t *testing.T, orders *fakeRunningOrderRepository, history *fakeHistoryRepository, version int) error {
		t.Helper()
		races := &fakeRaceRepository{stored: teamRace(t, teamID, distanceParams(t))}
		useCase := NewConfirmRunningOrderUseCase(&fakeUnitOfWork{}, races, entriesOf(nil), orders, history, runningorder.HistoryFactory{}, clock, teamID)
		return useCase.Execute(context.Background(), organizer(), RunningOrderCommandFactory{}.Confirm("r1", "e1", version))
	}

	t.Run("confirms the order and keeps a history entry", func(t *testing.T) {
		orders, history := ordersOf(storedOrder(runningorder.Draft, 2, member("m1"), member("m2"))), &fakeHistoryRepository{}
		if err := confirm(t, orders, history, 2); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(orders.saved) != 1 || orders.saved[0].Status() != runningorder.Confirmed || !orders.savedInTransaction {
			t.Errorf("saved %+v, want the confirmed order in the transaction", orders.saved)
		}
		if len(history.appended) != 1 || history.appended[0].Status() != runningorder.Confirmed || history.appended[0].ChangedBy() != "u2" {
			t.Errorf("history = %+v, want the confirmed order by u2", history.appended)
		}
	})

	failures := []struct {
		name    string
		orders  *fakeRunningOrderRepository
		version int
		check   func(error) bool
	}{
		{name: "an empty leg", orders: ordersOf(storedOrder(runningorder.Draft, 2, member("m1"), nil)), version: 2, check: isValidation},
		{name: "a confirmed order", orders: ordersOf(storedOrder(runningorder.Confirmed, 2, member("m1"), member("m2"))), version: 2, check: isState},
		{name: "a stale version", orders: ordersOf(storedOrder(runningorder.Draft, 2, member("m1"), member("m2"))), version: 1, check: isConflict},
		{name: "an entry without an order", orders: ordersOf(), version: 0, check: isNotFound},
	}
	for _, testCase := range failures {
		t.Run("rejects "+testCase.name, func(t *testing.T) {
			history := &fakeHistoryRepository{}
			if err := confirm(t, testCase.orders, history, testCase.version); !testCase.check(err) {
				t.Errorf("Execute() error = %v", err)
			}
			if len(testCase.orders.saved) != 0 || len(history.appended) != 0 {
				t.Error("the running order was saved")
			}
		})
	}
}
