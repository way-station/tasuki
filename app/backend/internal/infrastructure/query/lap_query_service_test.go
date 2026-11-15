package query

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/laps"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres"
)

var errDatabase = errors.New("database is down")

type fakeRunnerReader struct {
	runners []postgres.Runner
	err     error
}

func (fake *fakeRunnerReader) IsConfirmedRunner(_ context.Context, _ shared.TeamID, _ shared.RaceID, memberID shared.MemberID) (bool, error) {
	for _, runner := range fake.runners {
		if runner.MemberID == memberID {
			return true, nil
		}
	}
	return false, fake.err
}

func (fake *fakeRunnerReader) ConfirmedRunners(context.Context, shared.TeamID, shared.RaceID) ([]postgres.Runner, error) {
	return fake.runners, fake.err
}

type fakeLapRecordReader struct {
	byMember map[shared.MemberID][]laps.LapRecord
	err      error
}

func (fake *fakeLapRecordReader) RaceLaps(context.Context, shared.RaceID) (map[shared.MemberID][]laps.LapRecord, error) {
	return fake.byMember, fake.err
}

func (fake *fakeLapRecordReader) MemberLaps(_ context.Context, _ shared.RaceID, memberID shared.MemberID) ([]laps.LapRecord, error) {
	return fake.byMember[memberID], fake.err
}

type fakeProfiles struct{ err error }

func (fake *fakeProfiles) GetProfile(_ context.Context, memberID shared.MemberID) (laps.ProfileRecord, error) {
	if fake.err != nil {
		return laps.ProfileRecord{}, fake.err
	}
	average, err := shared.PaceFromSeconds(208)
	if err != nil {
		return laps.ProfileRecord{}, err
	}
	return laps.ProfileRecord{Name: string(memberID), Average: average, LapCount: 8}, nil
}

func lapRecord(t *testing.T, id shared.LapID, seconds int) laps.LapRecord {
	t.Helper()
	lapTime, err := shared.PaceFromSeconds(seconds)
	if err != nil {
		t.Fatal(err)
	}
	return laps.LapRecord{ID: id, Time: lapTime}
}

func TestLapQueryService(t *testing.T) {
	runners := []postgres.Runner{{MemberID: "m2", Name: "Hanako"}, {MemberID: "m1", Name: "Taro"}}

	t.Run("lists the runners in their order, with empty laps for runners without any", func(t *testing.T) {
		records := &fakeLapRecordReader{byMember: map[shared.MemberID][]laps.LapRecord{"m1": {lapRecord(t, "l1", 200)}, "m9": {lapRecord(t, "l9", 300)}}}
		service := NewLapQueryService(&fakeRunnerReader{runners: runners}, records, &fakeProfiles{})
		listed, err := service.ListRaceLaps(context.Background(), "t1", "r1")
		if err != nil {
			t.Fatalf("ListRaceLaps() error = %v", err)
		}
		if len(listed) != 2 || listed[0].MemberID != "m2" || listed[0].Laps == nil || len(listed[0].Laps) != 0 {
			t.Fatalf("ListRaceLaps() = %+v, want Hanako without laps first", listed)
		}
		if listed[1].MemberID != "m1" || listed[1].Name != "Taro" || len(listed[1].Laps) != 1 || listed[1].Laps[0].ID != "l1" {
			t.Errorf("listed[1] = %+v, want Taro with l1", listed[1])
		}
	})

	t.Run("returns the laps with the average", func(t *testing.T) {
		records := &fakeLapRecordReader{byMember: map[shared.MemberID][]laps.LapRecord{"m1": {lapRecord(t, "l1", 200)}}}
		service := NewLapQueryService(&fakeRunnerReader{}, records, &fakeProfiles{})
		myLaps, err := service.GetMyLaps(context.Background(), "r1", "m1")
		if err != nil || len(myLaps.Laps) != 1 || myLaps.Average.String() != "3:28" || myLaps.LapCount != 8 {
			t.Errorf("GetMyLaps() = %+v, %v, want l1 with 3:28 over 8 laps", myLaps, err)
		}
	})

	t.Run("asks PostgreSQL who the confirmed runners are", func(t *testing.T) {
		service := NewLapQueryService(&fakeRunnerReader{runners: runners}, &fakeLapRecordReader{}, &fakeProfiles{})
		if runner, err := service.IsConfirmedRunner(context.Background(), "t1", "r1", "m1"); err != nil || !runner {
			t.Errorf("IsConfirmedRunner() = %v, %v, want true", runner, err)
		}
	})

	tests := []struct {
		name    string
		runners *fakeRunnerReader
		records *fakeLapRecordReader
		want    error
	}{
		{name: "PostgreSQL", runners: &fakeRunnerReader{err: errDatabase}, records: &fakeLapRecordReader{}, want: errDatabase},
		{name: "MongoDB", runners: &fakeRunnerReader{runners: runners}, records: &fakeLapRecordReader{err: errDatabase}, want: errDatabase},
		{name: "a race of another team", runners: &fakeRunnerReader{err: fmt.Errorf("race r1: %w", shared.ErrNotFound)}, records: &fakeLapRecordReader{}, want: shared.ErrNotFound},
	}
	for _, testCase := range tests {
		t.Run("returns the error of "+testCase.name, func(t *testing.T) {
			service := NewLapQueryService(testCase.runners, testCase.records, &fakeProfiles{})
			if _, err := service.ListRaceLaps(context.Background(), "t1", "r1"); !errors.Is(err, testCase.want) {
				t.Errorf("ListRaceLaps() error = %v, want %v", err, testCase.want)
			}
		})
	}
}
