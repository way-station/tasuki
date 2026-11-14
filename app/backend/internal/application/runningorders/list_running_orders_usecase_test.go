package runningorders

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fakeRunningOrderQueryService struct {
	record         RaceRunningOrdersRecord
	err            error
	rosterMemberID *shared.MemberID
}

func (fake *fakeRunningOrderQueryService) ListRunningOrders(_ context.Context, _ shared.TeamID, _ shared.RaceID, rosterMemberID *shared.MemberID) (RaceRunningOrdersRecord, error) {
	fake.rosterMemberID = rosterMemberID
	return fake.record, fake.err
}

// raceRecord returns the distance race with a confirmed order of Team A, where m1 runs
// leg 1, and a draft of Team B.
func raceRecord(t *testing.T, answer race.Answer) RaceRunningOrdersRecord {
	t.Helper()
	taro := &RunnerRecord{MemberID: "m1", Name: "Taro", AverageSeconds: 220}
	return RaceRunningOrdersRecord{
		Description: distanceParams(t),
		Answer:      answer,
		RunningOrders: []RunningOrderRecord{
			{EntryID: "e1", EntryName: "Team A", Status: runningorder.Confirmed, Version: 3, Legs: []LegRecord{{Number: 1, Runner: taro}, {Number: 2}}},
			{EntryID: "e2", EntryName: "Team B", Status: runningorder.Draft, Version: 1, Legs: []LegRecord{{Number: 1}, {Number: 2}}},
		},
	}
}

func listOrders(t *testing.T, caller auth.MeDTO, record RaceRunningOrdersRecord) ([]RunningOrderDTO, *fakeRunningOrderQueryService, error) {
	t.Helper()
	query := &fakeRunningOrderQueryService{record: record}
	got, err := NewListRunningOrdersUseCase(query, race.VisibilityPolicy{}, runningorder.VisibilityPolicy{}, Predictor{}, RunningOrderDTOFactory{}, teamID).Execute(context.Background(), caller, "r1")
	return got, query, err
}

func TestListRunningOrdersUseCase(t *testing.T) {
	t.Run("returns drafts and versions to an organizer", func(t *testing.T) {
		got, query, err := listOrders(t, organizer(), raceRecord(t, race.NotInvited))
		if err != nil || len(got) != 2 || query.rosterMemberID != nil {
			t.Fatalf("Execute = %+v, %v (read as %v), want both orders as nobody", got, err, query.rosterMemberID)
		}
		if version, ok := got[1].Version(); !ok || version != 1 || got[1].Status() != runningorder.Draft {
			t.Errorf("second order = %+v, want the draft with version 1", got[1])
		}
		if _, ok := got[1].Prediction(); ok {
			t.Error("an order without runners has a prediction")
		}
	})

	t.Run("returns confirmed orders without versions to an attending member", func(t *testing.T) {
		got, query, err := listOrders(t, linkedMember(), raceRecord(t, race.Attending))
		if err != nil || len(got) != 1 || query.rosterMemberID == nil || *query.rosterMemberID != "m1" {
			t.Fatalf("Execute = %+v, %v, want the confirmed order read as m1", got, err)
		}
		order := got[0]
		if _, ok := order.Version(); ok || order.EntryID() != "e1" || order.EntryName() != "Team A" {
			t.Errorf("order = %+v, want Team A without a version", order)
		}
		legs := order.Legs()
		first, hasRunner := legs[0].Runner()
		distance, hasDistance := legs[1].DistanceKm()
		if len(legs) != 2 || !hasRunner || first.MemberID() != "m1" || first.Name() != "Taro" || !hasDistance || distance != 3.2 {
			t.Errorf("legs = %+v, want Taro on leg 1 and 3.2 km for leg 2", legs)
		}
		if _, ok := legs[1].Runner(); ok {
			t.Error("leg 2 has a runner")
		}
		prediction, ok := order.Prediction()
		if total, hasTotal := prediction.TotalTime(); !ok || !hasTotal || total.String() != "0:18:20" || !prediction.HasEmptyLegs() {
			t.Errorf("prediction = %+v, want 0:18:20 with empty legs", prediction)
		}
	})

	t.Run("returns an empty list to a member who has not answered attending", func(t *testing.T) {
		if got, _, err := listOrders(t, linkedMember(), raceRecord(t, race.Undecided)); err != nil || len(got) != 0 {
			t.Errorf("Execute = %+v, %v, want an empty list", got, err)
		}
	})

	for _, answer := range []race.Answer{race.NotInvited, race.NotAttending} {
		t.Run("returns not found to a member who may not see the race", func(t *testing.T) {
			if _, _, err := listOrders(t, linkedMember(), raceRecord(t, answer)); !errors.Is(err, shared.ErrNotFound) {
				t.Errorf("Execute error = %v, want %v", err, shared.ErrNotFound)
			}
		})
	}

	t.Run("returns the error of a race the team does not have", func(t *testing.T) {
		query := &fakeRunningOrderQueryService{err: shared.ErrNotFound}
		_, err := NewListRunningOrdersUseCase(query, race.VisibilityPolicy{}, runningorder.VisibilityPolicy{}, Predictor{}, RunningOrderDTOFactory{}, teamID).Execute(context.Background(), organizer(), "r1")
		if !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Execute error = %v, want %v", err, shared.ErrNotFound)
		}
	})
}

func TestGetRunningOrderUseCase(t *testing.T) {
	query := &fakeRunningOrderQueryService{record: raceRecord(t, race.Attending)}
	useCase := NewGetRunningOrderUseCase(NewListRunningOrdersUseCase(query, race.VisibilityPolicy{}, runningorder.VisibilityPolicy{}, Predictor{}, RunningOrderDTOFactory{}, teamID))

	if got, err := useCase.Execute(context.Background(), organizer(), "r1", "e2"); err != nil || got.EntryID() != "e2" {
		t.Errorf("Execute(e2) = %+v, %v, want the order of e2", got, err)
	}
	if _, err := useCase.Execute(context.Background(), linkedMember(), "r1", "e2"); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Execute(draft for a member) error = %v, want %v", err, shared.ErrNotFound)
	}
}
