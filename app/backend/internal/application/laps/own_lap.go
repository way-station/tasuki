package laps

import (
	"context"
	"fmt"
	"slices"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// requireLap returns the roster member the caller acts as when they may change their laps
// for the race and have the lap. It returns lap.ErrNotRunner when the caller is not a runner
// in a confirmed running order of the race, and an error wrapping shared.ErrNotFound when
// they have no such lap in the race.
func requireLap(ctx context.Context, query LapQueryService, records lap.Repository, policy lap.SubmissionPolicy, teamID shared.TeamID, caller auth.MeDTO, raceID shared.RaceID, lapID shared.LapID) (shared.MemberID, error) {
	memberID, runner, err := confirmedRunner(ctx, query, teamID, caller, raceID)
	if err != nil {
		return "", err
	}
	if err := policy.CanChange(runner); err != nil {
		return "", err
	}
	record, err := records.Find(ctx, raceID, memberID)
	if err != nil {
		return "", err
	}
	if !slices.ContainsFunc(record.Laps(), func(stored lap.Lap) bool { return stored.ID() == lapID }) {
		return "", fmt.Errorf("lap %s of race %s: %w", lapID, raceID, shared.ErrNotFound)
	}
	return memberID, nil
}
