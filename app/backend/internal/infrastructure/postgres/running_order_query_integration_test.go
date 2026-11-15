//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

func TestRunningOrderQueryService(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	ctx := context.Background()
	fixture := seedRunningOrderRace(t, client, testTeamID, "Taro", "Jiro")
	taro := fixture.members[0]
	factory := runningorder.NewFactory(random.IDGenerator{})
	saved, err := factory.New(fixture.entryID, legsOf(&taro, nil), runningorder.Shape{LegCount: 2}, fixture.members)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewRunningOrderRepository(client, factory).Save(ctx, saved); err != nil {
		t.Fatal(err)
	}
	queryService := NewRunningOrderQueryService(client)

	t.Run("returns the race, the answer, and the legs with the runners' average times", func(t *testing.T) {
		record, err := queryService.ListRunningOrders(ctx, testTeamID, fixture.raceID, &taro)
		if err != nil {
			t.Fatalf("ListRunningOrders: %v", err)
		}
		if record.Answer != race.Attending || record.Description.Type != race.Distance || len(record.Description.LegDistancesKm) != 2 {
			t.Errorf("race = %+v, want the distance race answered attending", record)
		}
		if len(record.RunningOrders) != 1 {
			t.Fatalf("running orders = %+v, want one", record.RunningOrders)
		}
		order := record.RunningOrders[0]
		if order.EntryID != fixture.entryID || order.EntryName != "Team A" || order.Status != runningorder.Draft || order.Version != 0 || len(order.Legs) != 2 {
			t.Fatalf("order = %+v, want the draft of Team A with two legs", order)
		}
		runner := order.Legs[0].Runner
		if order.Legs[0].Number != 1 || runner == nil || runner.MemberID != taro || runner.Name != "Taro" || runner.AverageSeconds != 210 {
			t.Errorf("leg 1 = %+v, runner %+v, want Taro at 210 s/km", order.Legs[0], runner)
		}
		if order.Legs[1].Number != 2 || order.Legs[1].Runner != nil {
			t.Errorf("leg 2 = %+v, want an empty leg", order.Legs[1])
		}
	})

	t.Run("reads nobody's answer without a roster member", func(t *testing.T) {
		if record, err := queryService.ListRunningOrders(ctx, testTeamID, fixture.raceID, nil); err != nil || record.Answer != race.NotInvited {
			t.Errorf("ListRunningOrders(nil) = %+v, %v, want not invited", record, err)
		}
	})

	t.Run("returns not found for a race of another team", func(t *testing.T) {
		const otherTeamID shared.TeamID = "01900000-0000-7000-8000-000000000002"
		if _, err := queryService.ListRunningOrders(ctx, otherTeamID, fixture.raceID, nil); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("ListRunningOrders error = %v, want ErrNotFound", err)
		}
	})
}
