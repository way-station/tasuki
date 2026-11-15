package runningorders

import (
	"context"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestRemoveRunnerUseCase(t *testing.T) {
	tests := []struct {
		name       string
		orders     *fakeRunningOrderRepository
		wantSaved  bool
		removeLaps bool
	}{
		{name: "empties the legs of a draft and keeps the laps", orders: ordersOf(storedOrder(runningorder.Draft, 1, member("m1"), member("m2"))), wantSaved: true},
		{name: "empties the legs of a confirmed order and removes the laps", orders: ordersOf(storedOrder(runningorder.Confirmed, 3, member("m1"), member("m2"))), wantSaved: true, removeLaps: true},
		{name: "removes the laps when the legs were already emptied", orders: ordersOf(storedOrder(runningorder.Confirmed, 3, nil, member("m2"))), removeLaps: true},
		{name: "does nothing without a running order", orders: ordersOf()},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			laps := &fakeLapRemover{}
			if err := NewRemoveRunnerUseCase(entriesOf(nil), testCase.orders, laps).RemoveRunner(context.Background(), "r1", "m1"); err != nil {
				t.Fatalf("RemoveRunner() error = %v", err)
			}
			if saved := len(testCase.orders.saved) == 1; saved != testCase.wantSaved {
				t.Fatalf("saved %+v, want saved = %v", testCase.orders.saved, testCase.wantSaved)
			}
			if testCase.wantSaved {
				if runner, ok := testCase.orders.saved[0].Legs()[0].Runner(); ok {
					t.Errorf("leg 1 runner = %s, want empty", runner)
				}
			}
			removed := len(laps.removed) == 1 && slices.Equal(laps.removed[0].memberIDs, []shared.MemberID{"m1"})
			if removed != testCase.removeLaps {
				t.Errorf("removed laps %+v, want removed = %v", laps.removed, testCase.removeLaps)
			}
		})
	}
}
