package laps

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func clock() time.Time { return testNow }

type transactionKey struct{}

func inTransaction(ctx context.Context) bool {
	marked, _ := ctx.Value(transactionKey{}).(bool)
	return marked
}

type fakeUnitOfWork struct{ err error }

func (fake *fakeUnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	fake.err = work(context.WithValue(ctx, transactionKey{}, true))
	return fake.err
}

// fakeOutbox keeps the entries that were added in a unit of work.
type fakeOutbox struct {
	entries []outbox.Entry
}

func (fake *fakeOutbox) Add(ctx context.Context, entry outbox.Entry) error {
	if !inTransaction(ctx) {
		return errors.New("the entry was added outside a unit of work")
	}
	fake.entries = append(fake.entries, entry)
	return nil
}

// onlySteps returns the steps of the one added entry.
func (fake *fakeOutbox) onlySteps(t *testing.T) []outbox.Step {
	t.Helper()
	if len(fake.entries) != 1 {
		t.Fatalf("entries = %d, want 1", len(fake.entries))
	}
	return fake.entries[0].Steps()
}

func submitCommand(t *testing.T, raceID shared.RaceID, times ...string) SubmitLapsCommand {
	t.Helper()
	command, err := LapCommandFactory{}.Submit(raceID, times)
	if err != nil {
		t.Fatal(err)
	}
	return command
}

func TestSubmitLapsUseCase(t *testing.T) {
	newUseCase := func(query *fakeLapQueryService, laps *fakeOutbox) *SubmitLapsUseCase {
		return NewSubmitLapsUseCase(&fakeUnitOfWork{}, query, laps, lap.SubmissionPolicy{}, lapFactory(), outbox.EntryFactory{}, clock, teamID)
	}

	t.Run("adds the laps and then recalculates the average", func(t *testing.T) {
		laps := &fakeOutbox{}
		err := newUseCase(&fakeLapQueryService{runners: []shared.MemberID{"m1"}}, laps).Execute(context.Background(), runnerCaller(), submitCommand(t, "r1", "3:20", "3:25"))
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		steps := laps.onlySteps(t)
		want := AddLapsStep{RaceID: "r1", MemberID: "m1", Laps: []StepLap{{ID: "l1", Seconds: 200, SubmittedAt: testNow}, {ID: "l2", Seconds: 205, SubmittedAt: testNow}}}
		if len(steps) != 2 {
			t.Fatalf("steps = %v, want 2", steps)
		}
		if added, ok := steps[0].(AddLapsStep); !ok || added.RaceID != want.RaceID || added.MemberID != want.MemberID || len(added.Laps) != 2 || added.Laps[0] != want.Laps[0] || added.Laps[1] != want.Laps[1] {
			t.Errorf("steps[0] = %+v, want %+v", steps[0], want)
		}
		if steps[1] != (RecalculateAverageStep{MemberID: "m1"}) {
			t.Errorf("steps[1] = %+v, want the average of m1", steps[1])
		}
	})

	t.Run("rejects a caller who is not a confirmed runner", func(t *testing.T) {
		laps := &fakeOutbox{}
		err := newUseCase(&fakeLapQueryService{}, laps).Execute(context.Background(), runnerCaller(), submitCommand(t, "r1", "3:20"))
		if !errors.Is(err, lap.ErrNotRunner) || len(laps.entries) != 0 {
			t.Errorf("Execute() error, entries = %v, %d, want lap.ErrNotRunner and none", err, len(laps.entries))
		}
	})

	t.Run("returns the error of the query service", func(t *testing.T) {
		laps := &fakeOutbox{}
		err := newUseCase(&fakeLapQueryService{err: errDatabase}, laps).Execute(context.Background(), runnerCaller(), submitCommand(t, "r1", "3:20"))
		if !errors.Is(err, errDatabase) || len(laps.entries) != 0 {
			t.Errorf("Execute() error, entries = %v, %d, want %v and none", err, len(laps.entries), errDatabase)
		}
	})
}
