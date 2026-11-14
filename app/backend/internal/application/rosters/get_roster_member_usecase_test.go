package rosters

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestGetRosterMemberUseCase(t *testing.T) {
	lapTime, err := shared.ParsePace("3:20")
	if err != nil {
		t.Fatal(err)
	}
	query := &fakeRosterQueryService{detail: RosterMemberDetailRecord{MemberID: "m1", Name: "Taro", Average: lapTime, LapCount: 1,
		LapsByRace: []RaceLapsRecord{{RaceID: "r1", RaceName: "Hakone", Laps: []LapRecord{{ID: "l1", Time: lapTime}}}}}}

	member, err := NewGetRosterMemberUseCase(query, RosterDTOFactory{}, teamID).Execute(context.Background(), "m1")
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	races := member.LapsByRace()
	if member.Name() != "Taro" || member.Average().String() != "3:20" || member.LapCount() != 1 || len(races) != 1 || races[0].RaceName() != "Hakone" || races[0].Laps()[0].ID() != "l1" {
		t.Errorf("Execute() = %+v, want Taro with the lap of Hakone", member)
	}

	if _, err := NewGetRosterMemberUseCase(query, RosterDTOFactory{}, teamID).Execute(context.Background(), "m9"); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Execute(m9) error = %v, want shared.ErrNotFound", err)
	}
}
