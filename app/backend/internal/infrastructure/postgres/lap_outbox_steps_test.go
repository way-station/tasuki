package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/laps"
	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

// stepLapRepository keeps the one record that the steps save.
type stepLapRepository struct{ saved *lap.Record }

func (repository *stepLapRepository) Find(_ context.Context, raceID shared.RaceID, memberID shared.MemberID) (*lap.Record, error) {
	if repository.saved == nil {
		return nil, fmt.Errorf("lap record of %s in %s: %w", memberID, raceID, shared.ErrNotFound)
	}
	return repository.saved, nil
}

func (repository *stepLapRepository) ListByMember(context.Context, shared.MemberID) ([]*lap.Record, error) {
	if repository.saved == nil {
		return nil, nil
	}
	return []*lap.Record{repository.saved}, nil
}

func (repository *stepLapRepository) Save(_ context.Context, record *lap.Record) error {
	repository.saved = record
	return nil
}

func (repository *stepLapRepository) Delete(context.Context, shared.RaceID, shared.MemberID) error {
	return nil
}

type stepMemberRepository struct{ member *roster.Member }

func (repository *stepMemberRepository) FindByID(context.Context, shared.MemberID) (*roster.Member, error) {
	return repository.member, nil
}

func (repository *stepMemberRepository) Save(_ context.Context, member *roster.Member) error {
	repository.member = member
	return nil
}

// runStored encodes the step as the outbox stores it, decodes it, and runs it.
func runStored(t *testing.T, steps map[string]OutboxStep, step outbox.Step) {
	t.Helper()
	registered, ok := steps[step.Kind()]
	if !ok {
		t.Fatalf("step %q is not registered", step.Kind())
	}
	payload, err := json.Marshal(step)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := registered.Decode(payload)
	if err != nil {
		t.Fatalf("Decode(%s): %v", payload, err)
	}
	if err := registered.Run(context.Background(), decoded); err != nil {
		t.Fatalf("Run(%s): %v", step.Kind(), err)
	}
}

func TestLapOutboxSteps(t *testing.T) {
	records := &stepLapRepository{}
	members := &stepMemberRepository{member: roster.NewMemberFactory(random.IDGenerator{}).Reconstruct("m1", "t1", "Taro", nil, roster.DefaultAverage(), 0)}
	runner := laps.NewLapStepRunner(records, members, lap.NewFactory(random.IDGenerator{}), lap.RecordFactory{}, roster.AverageService{})
	steps := map[string]OutboxStep{}
	for _, step := range LapOutboxSteps(runner) {
		steps[step.Kind] = step
	}
	submittedAt := time.Date(2026, 11, 11, 9, 0, 0, 0, time.UTC)

	t.Run("only recalculating the average writes to PostgreSQL", func(t *testing.T) {
		for kind, step := range steps {
			if step.FinalPostgreSQL != (kind == laps.RecalculateAverageKind) {
				t.Errorf("%s: FinalPostgreSQL = %v", kind, step.FinalPostgreSQL)
			}
		}
		if len(steps) != 5 {
			t.Errorf("steps = %d, want 5", len(steps))
		}
	})

	t.Run("runs the stored steps with the runner", func(t *testing.T) {
		runStored(t, steps, laps.AddLapsStep{RaceID: "r1", MemberID: "m1", Laps: []laps.StepLap{{ID: "l1", Seconds: 200, SubmittedAt: submittedAt}, {ID: "l2", Seconds: 220, SubmittedAt: submittedAt}}})
		runStored(t, steps, laps.EditLapStep{RaceID: "r1", MemberID: "m1", LapID: "l2", Seconds: 210})
		runStored(t, steps, laps.DeleteLapStep{RaceID: "r1", MemberID: "m1", LapID: "l1"})
		runStored(t, steps, laps.RecalculateAverageStep{MemberID: "m1"})
		stored := records.saved.Laps()
		if len(stored) != 1 || stored[0].ID() != "l2" || stored[0].Time().String() != "3:30" || !stored[0].SubmittedAt().Equal(submittedAt) {
			t.Errorf("laps = %v, want l2 at 3:30", stored)
		}
		if members.member.Average().String() != "3:30" || members.member.LapCount() != 1 {
			t.Errorf("average = %s over %d laps, want 3:30 over 1", members.member.Average(), members.member.LapCount())
		}
	})

	t.Run("rejects a step of another type", func(t *testing.T) {
		if err := steps[laps.AddLapsKind].Run(context.Background(), laps.RecalculateAverageStep{MemberID: "m1"}); err == nil {
			t.Error("Run with another step type succeeded")
		}
	})
}
