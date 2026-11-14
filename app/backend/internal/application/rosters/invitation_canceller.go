package rosters

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// InvitationCanceller cancels a member's invitation to a race, with the side effects of
// cancelling an invitation (docs/design/api.md). It is called in the unit of work that marks
// the member as left, and is implemented by the participation Application layer (#91).
type InvitationCanceller interface {
	CancelInvitation(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error
}
