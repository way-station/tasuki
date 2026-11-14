package laps

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LapQueryService reads the laps of a race with what the lap screens show. It is
// implemented in the infrastructure layer.
type LapQueryService interface {
	// IsConfirmedRunner reports whether the roster member runs a leg of a confirmed running
	// order of the race. It is false when the team has no such race.
	IsConfirmedRunner(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID, memberID shared.MemberID) (bool, error)
	// GetMyLaps returns the member's laps for the race in submission order, with their
	// average time and lap count over every race.
	GetMyLaps(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) (MyLapsRecord, error)
	// ListRaceLaps returns every runner of the confirmed running orders of the race, by name,
	// with their laps for the race in submission order; a runner with no laps has none. It
	// returns an error wrapping shared.ErrNotFound when the team has no such race.
	ListRaceLaps(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]RunnerLapsRecord, error)
}

// MyLapsRecord is a member's laps for a race as GetMyLaps reads them.
type MyLapsRecord struct {
	Laps     []LapRecord
	Average  shared.Pace
	LapCount int
}

// RunnerLapsRecord is a runner with their laps as ListRaceLaps reads them.
type RunnerLapsRecord struct {
	MemberID shared.MemberID
	Name     string
	Laps     []LapRecord
}

// LapRecord is one lap.
type LapRecord struct {
	ID   shared.LapID
	Time shared.Pace
}

// ProfileQueryService reads the profile of a roster member. It is implemented in the
// infrastructure layer.
type ProfileQueryService interface {
	// GetProfile returns the member's name, average time, and lap count. It returns an error
	// wrapping shared.ErrNotFound when there is no such member.
	GetProfile(ctx context.Context, memberID shared.MemberID) (ProfileRecord, error)
}

// ProfileRecord is a roster member as GetProfile reads them.
type ProfileRecord struct {
	Name     string
	Average  shared.Pace
	LapCount int
}
