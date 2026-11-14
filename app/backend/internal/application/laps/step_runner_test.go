package laps

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

var testNow = time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)

type sequentialIDs struct{ next int }

func (ids *sequentialIDs) NewID() string {
	ids.next++
	return fmt.Sprintf("l%d", ids.next)
}

func lapFactory() lap.Factory { return lap.NewFactory(&sequentialIDs{}) }

// storedLap returns a lap read from storage.
func storedLap(t *testing.T, id shared.LapID, text string) lap.Lap {
	t.Helper()
	return lapFactory().Reconstruct(id, pace(t, text), testNow)
}

func recordOf(raceID shared.RaceID, memberID shared.MemberID, laps ...lap.Lap) *lap.Record {
	return lap.RecordFactory{}.Reconstruct(raceID, memberID, laps)
}

type recordKey struct {
	raceID   shared.RaceID
	memberID shared.MemberID
}

// fakeLapRepository keeps lap records by race and member, and counts the saves.
type fakeLapRepository struct {
	stored map[recordKey]*lap.Record
	saves  int
	err    error
}

func recordsOf(records ...*lap.Record) *fakeLapRepository {
	fake := &fakeLapRepository{stored: map[recordKey]*lap.Record{}}
	for _, record := range records {
		fake.stored[recordKey{raceID: record.RaceID(), memberID: record.MemberID()}] = record
	}
	return fake
}

func (fake *fakeLapRepository) Find(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) (*lap.Record, error) {
	if fake.err != nil {
		return nil, fake.err
	}
	record, ok := fake.stored[recordKey{raceID: raceID, memberID: memberID}]
	if !ok {
		return nil, fmt.Errorf("lap record of %s in %s: %w", memberID, raceID, shared.ErrNotFound)
	}
	// A copy, so that changes reach the fake only through Save.
	return lap.RecordFactory{}.Reconstruct(record.RaceID(), record.MemberID(), record.Laps()), nil
}

func (fake *fakeLapRepository) ListByMember(_ context.Context, memberID shared.MemberID) ([]*lap.Record, error) {
	if fake.err != nil {
		return nil, fake.err
	}
	var records []*lap.Record
	for key, record := range fake.stored {
		if key.memberID == memberID {
			records = append(records, record)
		}
	}
	return records, nil
}

func (fake *fakeLapRepository) Save(_ context.Context, record *lap.Record) error {
	fake.stored[recordKey{raceID: record.RaceID(), memberID: record.MemberID()}] = record
	fake.saves++
	return nil
}

func (fake *fakeLapRepository) Delete(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) error {
	if fake.err != nil {
		return fake.err
	}
	key := recordKey{raceID: raceID, memberID: memberID}
	if _, ok := fake.stored[key]; !ok {
		return fmt.Errorf("lap record of %s in %s: %w", memberID, raceID, shared.ErrNotFound)
	}
	delete(fake.stored, key)
	return nil
}

func (fake *fakeLapRepository) times(raceID shared.RaceID, memberID shared.MemberID) []string {
	record, ok := fake.stored[recordKey{raceID: raceID, memberID: memberID}]
	if !ok {
		return nil
	}
	var times []string
	for _, stored := range record.Laps() {
		times = append(times, string(stored.ID())+" "+stored.Time().String())
	}
	return times
}

type fakeMemberRepository struct {
	stored *roster.Member
	saved  *roster.Member
}

func (fake *fakeMemberRepository) FindByID(_ context.Context, id shared.MemberID) (*roster.Member, error) {
	if fake.stored == nil || fake.stored.ID() != id {
		return nil, fmt.Errorf("member %s: %w", id, shared.ErrNotFound)
	}
	return fake.stored, nil
}

func (fake *fakeMemberRepository) Save(_ context.Context, member *roster.Member) error {
	fake.saved = member
	return nil
}

func stepRunner(records *fakeLapRepository, members *fakeMemberRepository) *LapStepRunner {
	return NewLapStepRunner(records, members, lapFactory(), lap.RecordFactory{}, roster.AverageService{})
}

func TestLapStepRunnerAddLaps(t *testing.T) {
	step := AddLapsStep{RaceID: "r1", MemberID: "m1", Laps: []StepLap{{ID: "l2", Seconds: 205, SubmittedAt: testNow}, {ID: "l3", Seconds: 210, SubmittedAt: testNow}}}

	t.Run("creates the record for the first laps of the race", func(t *testing.T) {
		records := recordsOf()
		if err := stepRunner(records, &fakeMemberRepository{}).AddLaps(context.Background(), step); err != nil {
			t.Fatalf("AddLaps() error = %v", err)
		}
		if got := records.times("r1", "m1"); !slices.Equal(got, []string{"l2 3:25", "l3 3:30"}) {
			t.Errorf("laps = %v, want l2 3:25 and l3 3:30", got)
		}
	})

	t.Run("adds the laps after the stored ones", func(t *testing.T) {
		records := recordsOf(recordOf("r1", "m1", storedLap(t, "l1", "3:20")))
		if err := stepRunner(records, &fakeMemberRepository{}).AddLaps(context.Background(), step); err != nil {
			t.Fatalf("AddLaps() error = %v", err)
		}
		if got := records.times("r1", "m1"); !slices.Equal(got, []string{"l1 3:20", "l2 3:25", "l3 3:30"}) {
			t.Errorf("laps = %v, want l1, l2, and l3", got)
		}
	})

	t.Run("does nothing more when it runs twice", func(t *testing.T) {
		records := recordsOf()
		runner := stepRunner(records, &fakeMemberRepository{})
		for range 2 {
			if err := runner.AddLaps(context.Background(), step); err != nil {
				t.Fatalf("AddLaps() error = %v", err)
			}
		}
		if got := records.times("r1", "m1"); !slices.Equal(got, []string{"l2 3:25", "l3 3:30"}) || records.saves != 1 {
			t.Errorf("laps, saves = %v, %d, want l2 and l3 saved once", got, records.saves)
		}
	})

	t.Run("returns the error of the repository", func(t *testing.T) {
		records := recordsOf()
		records.err = errDatabase
		if err := stepRunner(records, &fakeMemberRepository{}).AddLaps(context.Background(), step); !errors.Is(err, errDatabase) {
			t.Errorf("AddLaps() error = %v, want %v", err, errDatabase)
		}
	})
}

func TestLapStepRunnerEditLap(t *testing.T) {
	step := EditLapStep{RaceID: "r1", MemberID: "m1", LapID: "l1", Seconds: 200}

	t.Run("sets the time of the lap", func(t *testing.T) {
		records := recordsOf(recordOf("r1", "m1", storedLap(t, "l1", "3:30"), storedLap(t, "l2", "3:25")))
		if err := stepRunner(records, &fakeMemberRepository{}).EditLap(context.Background(), step); err != nil {
			t.Fatalf("EditLap() error = %v", err)
		}
		if got := records.times("r1", "m1"); !slices.Equal(got, []string{"l1 3:20", "l2 3:25"}) {
			t.Errorf("laps = %v, want l1 3:20 and l2 3:25", got)
		}
	})

	tests := []struct {
		name    string
		records *fakeLapRepository
	}{
		{name: "the lap is gone", records: recordsOf(recordOf("r1", "m1", storedLap(t, "l2", "3:25")))},
		{name: "the record is gone", records: recordsOf()},
	}
	for _, testCase := range tests {
		t.Run("does nothing when "+testCase.name, func(t *testing.T) {
			if err := stepRunner(testCase.records, &fakeMemberRepository{}).EditLap(context.Background(), step); err != nil {
				t.Fatalf("EditLap() error = %v", err)
			}
			if testCase.records.saves != 0 {
				t.Errorf("saves = %d, want 0", testCase.records.saves)
			}
		})
	}
}

func TestLapStepRunnerDeleteLap(t *testing.T) {
	step := DeleteLapStep{RaceID: "r1", MemberID: "m1", LapID: "l1"}

	t.Run("deletes the lap", func(t *testing.T) {
		records := recordsOf(recordOf("r1", "m1", storedLap(t, "l1", "3:30"), storedLap(t, "l2", "3:25")))
		if err := stepRunner(records, &fakeMemberRepository{}).DeleteLap(context.Background(), step); err != nil {
			t.Fatalf("DeleteLap() error = %v", err)
		}
		if got := records.times("r1", "m1"); !slices.Equal(got, []string{"l2 3:25"}) {
			t.Errorf("laps = %v, want l2 3:25", got)
		}
	})

	t.Run("does nothing more when it runs twice", func(t *testing.T) {
		records := recordsOf(recordOf("r1", "m1", storedLap(t, "l1", "3:30"), storedLap(t, "l2", "3:25")))
		runner := stepRunner(records, &fakeMemberRepository{})
		for range 2 {
			if err := runner.DeleteLap(context.Background(), step); err != nil {
				t.Fatalf("DeleteLap() error = %v", err)
			}
		}
		if got := records.times("r1", "m1"); !slices.Equal(got, []string{"l2 3:25"}) || records.saves != 1 {
			t.Errorf("laps, saves = %v, %d, want l2 saved once", got, records.saves)
		}
	})
}

func TestLapStepRunnerRemoveRaceLaps(t *testing.T) {
	step := RemoveRaceLapsStep{RaceID: "r1", MemberID: "m1"}
	others := func() []*lap.Record {
		return []*lap.Record{recordOf("r2", "m1", storedLap(t, "l3", "3:40")), recordOf("r1", "m2", storedLap(t, "l4", "3:50"))}
	}

	t.Run("deletes the member's laps for the race and keeps the others", func(t *testing.T) {
		records := recordsOf(append(others(), recordOf("r1", "m1", storedLap(t, "l1", "3:30"), storedLap(t, "l2", "3:25")))...)
		if err := stepRunner(records, &fakeMemberRepository{}).RemoveRaceLaps(context.Background(), step); err != nil {
			t.Fatalf("RemoveRaceLaps() error = %v", err)
		}
		if got := records.times("r1", "m1"); got != nil {
			t.Errorf("laps of m1 in r1 = %v, want none", got)
		}
		if !slices.Equal(records.times("r2", "m1"), []string{"l3 3:40"}) || !slices.Equal(records.times("r1", "m2"), []string{"l4 3:50"}) {
			t.Errorf("the other records changed: %v", records.stored)
		}
	})

	t.Run("does nothing more when it runs twice", func(t *testing.T) {
		records := recordsOf(append(others(), recordOf("r1", "m1", storedLap(t, "l1", "3:30")))...)
		runner := stepRunner(records, &fakeMemberRepository{})
		for range 2 {
			if err := runner.RemoveRaceLaps(context.Background(), step); err != nil {
				t.Fatalf("RemoveRaceLaps() error = %v", err)
			}
		}
		if got := records.times("r1", "m1"); got != nil || len(records.stored) != 2 {
			t.Errorf("laps of m1 in r1 = %v, records = %d, want none and 2", got, len(records.stored))
		}
	})

	t.Run("returns a database error", func(t *testing.T) {
		records := recordsOf()
		records.err = errDatabase
		if err := stepRunner(records, &fakeMemberRepository{}).RemoveRaceLaps(context.Background(), step); !errors.Is(err, errDatabase) {
			t.Errorf("RemoveRaceLaps() error = %v, want %v", err, errDatabase)
		}
	})
}

func TestLapStepRunnerRecalculateAverage(t *testing.T) {
	member := func(t *testing.T) *roster.Member {
		t.Helper()
		return roster.NewMemberFactory(&sequentialIDs{}).Reconstruct("m1", teamID, "Taro", nil, pace(t, "3:30"), 5)
	}

	t.Run("recalculates from the laps of every race", func(t *testing.T) {
		records := recordsOf(
			recordOf("r1", "m1", storedLap(t, "l1", "3:20"), storedLap(t, "l2", "3:25")),
			recordOf("r2", "m1", storedLap(t, "l3", "3:25")),
			recordOf("r1", "m2", storedLap(t, "l4", "5:00")),
		)
		members := &fakeMemberRepository{stored: member(t)}
		if err := stepRunner(records, members).RecalculateAverage(context.Background(), RecalculateAverageStep{MemberID: "m1"}); err != nil {
			t.Fatalf("RecalculateAverage() error = %v", err)
		}
		// (3:20 + 3:25 + 3:25) ÷ 3 = 3:23.3 → 3:23 (docs/design/database.md)
		if members.saved == nil || members.saved.Average().String() != "3:23" || members.saved.LapCount() != 3 {
			t.Errorf("saved = %v, want 3:23 over 3 laps", members.saved)
		}
	})

	t.Run("goes back to the default without laps", func(t *testing.T) {
		members := &fakeMemberRepository{stored: member(t)}
		if err := stepRunner(recordsOf(), members).RecalculateAverage(context.Background(), RecalculateAverageStep{MemberID: "m1"}); err != nil {
			t.Fatalf("RecalculateAverage() error = %v", err)
		}
		if members.saved == nil || members.saved.Average() != roster.DefaultAverage() || members.saved.LapCount() != 0 {
			t.Errorf("saved = %v, want the default average over 0 laps", members.saved)
		}
	})
}
