package rosters

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// findTeamMember returns the roster member. It returns an error wrapping shared.ErrNotFound
// unless the member belongs to the team.
func findTeamMember(ctx context.Context, members roster.MemberRepository, teamID shared.TeamID, memberID shared.MemberID) (*roster.Member, error) {
	member, err := members.FindByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if member.TeamID() != teamID {
		return nil, fmt.Errorf("roster member %s: %w", memberID, shared.ErrNotFound)
	}
	return member, nil
}
