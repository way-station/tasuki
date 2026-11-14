package laps

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func editCommand(t *testing.T, raceID shared.RaceID, lapID shared.LapID, text string) EditLapCommand {
	t.Helper()
	command, err := LapCommandFactory{}.Edit(raceID, lapID, text)
	if err != nil {
		t.Fatal(err)
	}
	return command
}

func TestEditLapUseCase(t *testing.T) {
	newUseCase := func(query *fakeLapQueryService, records *fakeLapRepository, laps *fakeOutbox) *EditLapUseCase {
		return NewEditLapUseCase(&fakeUnitOfWork{}, query, records, laps, lap.SubmissionPolicy{}, outbox.EntryFactory{}, teamID)
	}
	runners := func() *fakeLapQueryService { return &fakeLapQueryService{runners: []shared.MemberID{"m1"}} }

	t.Run("edits the lap and then recalculates the average", func(t *testing.T) {
		records, laps := recordsOf(recordOf("r1", "m1", storedLap(t, "l1", "3:30"))), &fakeOutbox{}
		if err := newUseCase(runners(), records, laps).Execute(context.Background(), runnerCaller(), editCommand(t, "r1", "l1", "3:20")); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		steps := laps.onlySteps(t)
		if len(steps) != 2 || steps[0] != (EditLapStep{RaceID: "r1", MemberID: "m1", LapID: "l1", Seconds: 200}) || steps[1] != (RecalculateAverageStep{MemberID: "m1"}) {
			t.Errorf("steps = %+v, want the edit of l1 and the average of m1", steps)
		}
		if records.saves != 0 {
			t.Errorf("saves = %d, want 0: the step changes the record", records.saves)
		}
	})

	t.Run("rejects a caller who is not a confirmed runner", func(t *testing.T) {
		records, laps := recordsOf(recordOf("r1", "m1", storedLap(t, "l1", "3:30"))), &fakeOutbox{}
		err := newUseCase(&fakeLapQueryService{}, records, laps).Execute(context.Background(), runnerCaller(), editCommand(t, "r1", "l1", "3:20"))
		if !errors.Is(err, lap.ErrNotRunner) || len(laps.entries) != 0 {
			t.Errorf("Execute() error, entries = %v, %d, want lap.ErrNotRunner and none", err, len(laps.entries))
		}
	})

	tests := []struct {
		name    string
		records *fakeLapRepository
	}{
		{name: "a lap the caller does not have", records: recordsOf(recordOf("r1", "m1", storedLap(t, "l2", "3:30")))},
		{name: "a lap of another member", records: recordsOf(recordOf("r1", "m2", storedLap(t, "l1", "3:30")))},
	}
	for _, testCase := range tests {
		t.Run("rejects "+testCase.name, func(t *testing.T) {
			laps := &fakeOutbox{}
			err := newUseCase(runners(), testCase.records, laps).Execute(context.Background(), runnerCaller(), editCommand(t, "r1", "l1", "3:20"))
			if !errors.Is(err, shared.ErrNotFound) || len(laps.entries) != 0 {
				t.Errorf("Execute() error, entries = %v, %d, want shared.ErrNotFound and none", err, len(laps.entries))
			}
		})
	}
}
