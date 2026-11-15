package laps

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const teamID shared.TeamID = "t1"

var errDatabase = errors.New("database is down")

// runnerCaller returns a member linked to roster member m1.
func runnerCaller() auth.MeDTO {
	factory := auth.MeDTOFactory{}
	return factory.Create("u1", "Taro", false, identity.CallerApproved, factory.CreateRosterMember("m1", "Taro"))
}

// organizer returns an organizer without a link.
func organizer() auth.MeDTO {
	return auth.MeDTOFactory{}.Create("u2", "Hanako", true, identity.CallerNone, nil)
}

func pace(t *testing.T, text string) shared.Pace {
	t.Helper()
	parsed, err := shared.ParsePace(text)
	if err != nil {
		t.Fatal(err)
	}
	return parsed
}

// fakeLapQueryService answers for team t1. runners are the confirmed runners of race r1.
type fakeLapQueryService struct {
	runners []shared.MemberID
	myLaps  MyLapsRecord
	race    []RunnerLapsRecord
	err     error
}

func (fake *fakeLapQueryService) IsConfirmedRunner(_ context.Context, team shared.TeamID, raceID shared.RaceID, memberID shared.MemberID) (bool, error) {
	if fake.err != nil {
		return false, fake.err
	}
	if team != teamID || raceID != "r1" {
		return false, nil
	}
	for _, runner := range fake.runners {
		if runner == memberID {
			return true, nil
		}
	}
	return false, nil
}

func (fake *fakeLapQueryService) GetMyLaps(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) (MyLapsRecord, error) {
	if raceID != "r1" || memberID != "m1" {
		return MyLapsRecord{}, fmt.Errorf("laps of %s in %s were not expected to be read", memberID, raceID)
	}
	return fake.myLaps, nil
}

func (fake *fakeLapQueryService) ListRaceLaps(_ context.Context, team shared.TeamID, raceID shared.RaceID) ([]RunnerLapsRecord, error) {
	if fake.err != nil {
		return nil, fake.err
	}
	if team != teamID || raceID != "r1" {
		return nil, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	return fake.race, nil
}

func TestGetMyLapsUseCase(t *testing.T) {
	t.Run("returns the runner's laps with their average", func(t *testing.T) {
		query := &fakeLapQueryService{runners: []shared.MemberID{"m1"}, myLaps: MyLapsRecord{
			Laps:    []LapRecord{{ID: "l1", Time: pace(t, "3:20")}, {ID: "l2", Time: pace(t, "3:25")}},
			Average: pace(t, "3:28"), LapCount: 8,
		}}
		laps, err := NewGetMyLapsUseCase(query, LapDTOFactory{}, teamID).Execute(context.Background(), runnerCaller(), "r1")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		got := laps.Laps()
		if len(got) != 2 || got[0].ID() != "l1" || got[0].Time().String() != "3:20" || got[1].ID() != "l2" {
			t.Errorf("Laps() = %v, want l1 3:20 and l2", got)
		}
		if laps.Average().String() != "3:28" || laps.LapCount() != 8 {
			t.Errorf("Average(), LapCount() = %s, %d, want 3:28, 8", laps.Average(), laps.LapCount())
		}
	})

	tests := []struct {
		name   string
		caller auth.MeDTO
		raceID shared.RaceID
	}{
		{name: "a member who is not a confirmed runner", caller: runnerCaller(), raceID: "r2"},
		{name: "an organizer without a link", caller: organizer(), raceID: "r1"},
	}
	for _, testCase := range tests {
		t.Run("rejects "+testCase.name, func(t *testing.T) {
			query := &fakeLapQueryService{runners: []shared.MemberID{"m1"}}
			_, err := NewGetMyLapsUseCase(query, LapDTOFactory{}, teamID).Execute(context.Background(), testCase.caller, testCase.raceID)
			if !errors.Is(err, lap.ErrNotRunner) {
				t.Errorf("Execute() error = %v, want lap.ErrNotRunner", err)
			}
		})
	}

	t.Run("returns the error of the query service", func(t *testing.T) {
		query := &fakeLapQueryService{err: errDatabase}
		if _, err := NewGetMyLapsUseCase(query, LapDTOFactory{}, teamID).Execute(context.Background(), runnerCaller(), "r1"); !errors.Is(err, errDatabase) {
			t.Errorf("Execute() error = %v, want %v", err, errDatabase)
		}
	})
}
