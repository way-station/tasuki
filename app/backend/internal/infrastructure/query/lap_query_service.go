// Package query holds the query services that read from both PostgreSQL and MongoDB. Each
// database is read by its own package; this package only combines what they return
// (ADR-0011).
package query

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/laps"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres"
)

// RunnerReader reads the runners of the confirmed running orders from PostgreSQL. It is
// implemented by postgres.LapReader.
type RunnerReader interface {
	IsConfirmedRunner(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID, memberID shared.MemberID) (bool, error)
	ConfirmedRunners(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]postgres.Runner, error)
}

// LapRecordReader reads the laps from MongoDB. It is implemented by mongodb.LapRecordReader.
type LapRecordReader interface {
	RaceLaps(ctx context.Context, raceID shared.RaceID) (map[shared.MemberID][]laps.LapRecord, error)
	MemberLaps(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) ([]laps.LapRecord, error)
}

// LapQueryService implements laps.LapQueryService: the runners and the average times from
// PostgreSQL, and the laps from MongoDB.
type LapQueryService struct {
	runners  RunnerReader
	laps     LapRecordReader
	profiles laps.ProfileQueryService
}

var _ laps.LapQueryService = (*LapQueryService)(nil)

// NewLapQueryService returns a LapQueryService that reads through runners, records, and
// profiles.
func NewLapQueryService(runners RunnerReader, records LapRecordReader, profiles laps.ProfileQueryService) *LapQueryService {
	return &LapQueryService{runners: runners, laps: records, profiles: profiles}
}

// IsConfirmedRunner reports whether the roster member runs a leg of a confirmed running
// order of the race. It is false when the team has no such race.
func (service *LapQueryService) IsConfirmedRunner(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID, memberID shared.MemberID) (bool, error) {
	return service.runners.IsConfirmedRunner(ctx, teamID, raceID, memberID)
}

// GetMyLaps returns the member's laps for the race with their average time and lap count.
func (service *LapQueryService) GetMyLaps(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) (laps.MyLapsRecord, error) {
	records, err := service.laps.MemberLaps(ctx, raceID, memberID)
	if err != nil {
		return laps.MyLapsRecord{}, err
	}
	profile, err := service.profiles.GetProfile(ctx, memberID)
	if err != nil {
		return laps.MyLapsRecord{}, err
	}
	return laps.MyLapsRecord{Laps: records, Average: profile.Average, LapCount: profile.LapCount}, nil
}

// ListRaceLaps returns every runner of the confirmed running orders of the race, by name,
// with their laps for the race. It returns an error wrapping shared.ErrNotFound when the
// team has no such race.
func (service *LapQueryService) ListRaceLaps(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]laps.RunnerLapsRecord, error) {
	runners, err := service.runners.ConfirmedRunners(ctx, teamID, raceID)
	if err != nil {
		return nil, err
	}
	byMember, err := service.laps.RaceLaps(ctx, raceID)
	if err != nil {
		return nil, err
	}
	records := make([]laps.RunnerLapsRecord, 0, len(runners))
	for _, runner := range runners {
		runnerLaps := byMember[runner.MemberID]
		if runnerLaps == nil {
			runnerLaps = []laps.LapRecord{}
		}
		records = append(records, laps.RunnerLapsRecord{MemberID: runner.MemberID, Name: runner.Name, Laps: runnerLaps})
	}
	return records, nil
}
