package laps

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestListRaceLapsUseCase(t *testing.T) {
	t.Run("returns every runner with their laps", func(t *testing.T) {
		query := &fakeLapQueryService{race: []RunnerLapsRecord{
			{MemberID: "m2", Name: "Hanako", Laps: []LapRecord{{ID: "l1", Time: pace(t, "3:20")}}},
			{MemberID: "m1", Name: "Taro", Laps: nil},
		}}
		runners, err := NewListRaceLapsUseCase(query, lap.VisibilityPolicy{}, LapDTOFactory{}, teamID).Execute(context.Background(), organizer(), "r1")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(runners) != 2 || runners[0].MemberID() != "m2" || runners[0].Name() != "Hanako" || len(runners[0].Laps()) != 1 || runners[0].Laps()[0].Time().String() != "3:20" {
			t.Fatalf("Execute() = %+v, want Hanako with 3:20 first", runners)
		}
		if runners[1].MemberID() != "m1" || len(runners[1].Laps()) != 0 {
			t.Errorf("runners[1] = %+v, want Taro without laps", runners[1])
		}
	})

	t.Run("leaves out the laps of others for a caller who is not an organizer", func(t *testing.T) {
		query := &fakeLapQueryService{race: []RunnerLapsRecord{{MemberID: "m2", Name: "Hanako"}, {MemberID: "m1", Name: "Taro"}}}
		runners, err := NewListRaceLapsUseCase(query, lap.VisibilityPolicy{}, LapDTOFactory{}, teamID).Execute(context.Background(), runnerCaller(), "r1")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(runners) != 1 || runners[0].MemberID() != "m1" {
			t.Errorf("Execute() = %+v, want only m1", runners)
		}
	})

	t.Run("rejects a race of another team", func(t *testing.T) {
		_, err := NewListRaceLapsUseCase(&fakeLapQueryService{}, lap.VisibilityPolicy{}, LapDTOFactory{}, teamID).Execute(context.Background(), organizer(), "r2")
		if !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Execute() error = %v, want shared.ErrNotFound", err)
		}
	})
}
