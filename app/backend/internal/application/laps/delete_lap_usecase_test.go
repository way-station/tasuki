package laps

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestDeleteLapUseCase(t *testing.T) {
	newUseCase := func(query *fakeLapQueryService, records *fakeLapRepository, laps *fakeOutbox) *DeleteLapUseCase {
		return NewDeleteLapUseCase(&fakeUnitOfWork{}, query, records, laps, lap.SubmissionPolicy{}, outbox.EntryFactory{}, teamID)
	}
	runners := func() *fakeLapQueryService { return &fakeLapQueryService{runners: []shared.MemberID{"m1"}} }
	command := LapCommandFactory{}.Delete("r1", "l1")

	t.Run("deletes the lap and then recalculates the average", func(t *testing.T) {
		records, laps := recordsOf(recordOf("r1", "m1", storedLap(t, "l1", "3:30"))), &fakeOutbox{}
		if err := newUseCase(runners(), records, laps).Execute(context.Background(), runnerCaller(), command); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		steps := laps.onlySteps(t)
		if len(steps) != 2 || steps[0] != (DeleteLapStep{RaceID: "r1", MemberID: "m1", LapID: "l1"}) || steps[1] != (RecalculateAverageStep{MemberID: "m1"}) {
			t.Errorf("steps = %+v, want the deletion of l1 and the average of m1", steps)
		}
	})

	t.Run("rejects a caller who is not a confirmed runner", func(t *testing.T) {
		records, laps := recordsOf(recordOf("r1", "m1", storedLap(t, "l1", "3:30"))), &fakeOutbox{}
		err := newUseCase(&fakeLapQueryService{}, records, laps).Execute(context.Background(), runnerCaller(), command)
		if !errors.Is(err, lap.ErrNotRunner) || len(laps.entries) != 0 {
			t.Errorf("Execute() error, entries = %v, %d, want lap.ErrNotRunner and none", err, len(laps.entries))
		}
	})

	t.Run("rejects a lap the caller does not have", func(t *testing.T) {
		laps := &fakeOutbox{}
		err := newUseCase(runners(), recordsOf(), laps).Execute(context.Background(), runnerCaller(), command)
		if !errors.Is(err, shared.ErrNotFound) || len(laps.entries) != 0 {
			t.Errorf("Execute() error, entries = %v, %d, want shared.ErrNotFound and none", err, len(laps.entries))
		}
	})
}
