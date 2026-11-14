package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/application/laps"
	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
)

// LapOutboxSteps returns the outbox steps of the laps, which runner runs
// (docs/design/api.md). Recalculating the average is the final PostgreSQL step, so it runs
// in the transaction that deletes the entry.
func LapOutboxSteps(runner *laps.LapStepRunner) []OutboxStep {
	return []OutboxStep{
		{Kind: laps.AddLapsKind, Decode: decodeStep[laps.AddLapsStep], Run: runStep(runner.AddLaps)},
		{Kind: laps.EditLapKind, Decode: decodeStep[laps.EditLapStep], Run: runStep(runner.EditLap)},
		{Kind: laps.DeleteLapKind, Decode: decodeStep[laps.DeleteLapStep], Run: runStep(runner.DeleteLap)},
		{Kind: laps.RemoveRaceLapsKind, Decode: decodeStep[laps.RemoveRaceLapsStep], Run: runStep(runner.RemoveRaceLaps)},
		{Kind: laps.RecalculateAverageKind, Decode: decodeStep[laps.RecalculateAverageStep], Run: runStep(runner.RecalculateAverage), FinalPostgreSQL: true},
	}
}

// decodeStep reads a step of type Step from the payload that the outbox stored as JSON.
func decodeStep[Step outbox.Step](payload []byte) (outbox.Step, error) {
	var step Step
	if err := json.Unmarshal(payload, &step); err != nil {
		return nil, err
	}
	return step, nil
}

// runStep returns a Run that passes a step of type Step to run.
func runStep[Step outbox.Step](run func(ctx context.Context, step Step) error) func(ctx context.Context, step outbox.Step) error {
	return func(ctx context.Context, step outbox.Step) error {
		typed, ok := step.(Step)
		if !ok {
			return fmt.Errorf("outbox step %q is a %T", step.Kind(), step)
		}
		return run(ctx, typed)
	}
}
