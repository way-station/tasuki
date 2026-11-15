package runningorders

import (
	"context"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestAdjustRunningOrdersUseCase(t *testing.T) {
	oneLeg := distanceParams(t)
	oneLeg.LegDistancesKm = []float64{5}
	// Entry e2 of the race has no running order and is skipped.
	entries := entriesOf(nil, participation.NewEntryFactory(fixedIDs{}).Reconstruct("e2", "r1", "Team B", nil))

	t.Run("drops legs and removes the laps of a runner of a confirmed order who loses one", func(t *testing.T) {
		orders, laps := ordersOf(storedOrder(runningorder.Confirmed, 3, member("m1"), member("m2"))), &fakeLapRemover{}
		if err := NewAdjustRunningOrdersUseCase(entries, orders, laps).AdjustRunningOrders(context.Background(), "r1", description(t, oneLeg)); err != nil {
			t.Fatalf("AdjustRunningOrders() error = %v", err)
		}
		if len(orders.saved) != 1 || len(orders.saved[0].Legs()) != 1 || orders.saved[0].Status() != runningorder.Confirmed {
			t.Errorf("saved %+v, want the confirmed order with one leg", orders.saved)
		}
		if len(laps.removed) != 1 || !slices.Equal(laps.removed[0].memberIDs, []shared.MemberID{"m2"}) {
			t.Errorf("removed laps %+v, want m2's", laps.removed)
		}
	})

	t.Run("leaves an order whose legs do not change", func(t *testing.T) {
		orders, laps := ordersOf(storedOrder(runningorder.Draft, 1, member("m1"), nil)), &fakeLapRemover{}
		if err := NewAdjustRunningOrdersUseCase(entries, orders, laps).AdjustRunningOrders(context.Background(), "r1", description(t, timeParams(t))); err != nil {
			t.Fatalf("AdjustRunningOrders() error = %v", err)
		}
		if len(orders.saved) != 0 || len(laps.removed) != 0 {
			t.Errorf("saved %+v, removed laps %+v, want neither", orders.saved, laps.removed)
		}
	})
}
