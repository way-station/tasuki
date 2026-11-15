package laps

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// confirmedRunner returns the roster member the caller acts as, and whether they run a leg
// of a confirmed running order of the race. A caller without a roster member, such as an
// organizer without a link, is not a runner.
func confirmedRunner(ctx context.Context, query LapQueryService, teamID shared.TeamID, caller auth.MeDTO, raceID shared.RaceID) (shared.MemberID, bool, error) {
	rosterMember, ok := caller.RosterMember()
	if !ok {
		return "", false, nil
	}
	runner, err := query.IsConfirmedRunner(ctx, teamID, raceID, rosterMember.ID())
	if err != nil {
		return "", false, err
	}
	return rosterMember.ID(), runner, nil
}
