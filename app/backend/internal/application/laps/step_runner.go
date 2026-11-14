package laps

import (
	"context"
	"errors"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LapStepRunner runs the outbox steps of this package. The Infrastructure layer calls it
// for each kind; RecalculateAverage runs in the transaction that deletes the outbox entry.
// Every step has the same effect when it runs twice (ADR-0016).
type LapStepRunner struct {
	records       lap.Repository
	members       roster.MemberRepository
	lapFactory    lap.Factory
	recordFactory lap.RecordFactory
	averages      roster.AverageService
}

// NewLapStepRunner returns a LapStepRunner.
func NewLapStepRunner(records lap.Repository, members roster.MemberRepository, lapFactory lap.Factory, recordFactory lap.RecordFactory, averages roster.AverageService) *LapStepRunner {
	return &LapStepRunner{records: records, members: members, lapFactory: lapFactory, recordFactory: recordFactory, averages: averages}
}

// AddLaps adds the laps whose IDs are not in the record yet, creating the record for the
// first laps of the race.
func (runner *LapStepRunner) AddLaps(ctx context.Context, step AddLapsStep) error {
	record, err := runner.records.Find(ctx, step.RaceID, step.MemberID)
	if errors.Is(err, shared.ErrNotFound) {
		record, err = runner.recordFactory.New(step.RaceID, step.MemberID), nil
	}
	if err != nil {
		return err
	}
	stored := make(map[shared.LapID]bool, len(record.Laps()))
	for _, existing := range record.Laps() {
		stored[existing.ID()] = true
	}
	added := false
	for _, stepLap := range step.Laps {
		if stored[stepLap.ID] {
			continue
		}
		lapTime, err := shared.PaceFromSeconds(stepLap.Seconds)
		if err != nil {
			return fmt.Errorf("lap %s of the step: %w", stepLap.ID, err)
		}
		record.Add(runner.lapFactory.Reconstruct(stepLap.ID, lapTime, stepLap.SubmittedAt))
		added = true
	}
	if !added {
		return nil
	}
	return runner.records.Save(ctx, record)
}

// EditLap sets the time of the lap. A lap that is gone is left as it is.
func (runner *LapStepRunner) EditLap(ctx context.Context, step EditLapStep) error {
	lapTime, err := shared.PaceFromSeconds(step.Seconds)
	if err != nil {
		return fmt.Errorf("lap %s of the step: %w", step.LapID, err)
	}
	return runner.change(ctx, step.RaceID, step.MemberID, func(record *lap.Record) error {
		return record.Edit(step.LapID, lapTime)
	})
}

// DeleteLap deletes the lap. A lap that is gone is left as it is.
func (runner *LapStepRunner) DeleteLap(ctx context.Context, step DeleteLapStep) error {
	return runner.change(ctx, step.RaceID, step.MemberID, func(record *lap.Record) error {
		return record.Delete(step.LapID)
	})
}

// RemoveRaceLaps deletes the member's laps for the race. Laps that are already gone are
// left as they are.
func (runner *LapStepRunner) RemoveRaceLaps(ctx context.Context, step RemoveRaceLapsStep) error {
	err := runner.records.Delete(ctx, step.RaceID, step.MemberID)
	if errors.Is(err, shared.ErrNotFound) {
		return nil
	}
	return err
}

// change applies change to the member's record for the race and saves it. When the record
// or the lap is gone, nothing is saved.
func (runner *LapStepRunner) change(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID, change func(record *lap.Record) error) error {
	record, err := runner.records.Find(ctx, raceID, memberID)
	if errors.Is(err, shared.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	err = change(record)
	if errors.Is(err, shared.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return runner.records.Save(ctx, record)
}

// RecalculateAverage sets the member's average time and lap count from all of their laps in
// every race (docs/design/database.md).
func (runner *LapStepRunner) RecalculateAverage(ctx context.Context, step RecalculateAverageStep) error {
	member, err := runner.members.FindByID(ctx, step.MemberID)
	if err != nil {
		return err
	}
	records, err := runner.records.ListByMember(ctx, step.MemberID)
	if err != nil {
		return err
	}
	var all []shared.Pace
	for _, record := range records {
		all = append(all, record.Times()...)
	}
	runner.averages.Recalculate(member, all)
	return runner.members.Save(ctx, member)
}
