package participations

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LeaveInvitationCanceller cancels the invitations of a roster member who leaves the team.
// It implements rosters.InvitationCanceller (#103), which the roster calls in the unit of
// work that marks the member as left; CancelInvitationUseCase joins that unit of work.
type LeaveInvitationCanceller struct {
	cancel         *CancelInvitationUseCase
	commandFactory ParticipationCommandFactory
}

// NewLeaveInvitationCanceller returns a LeaveInvitationCanceller.
func NewLeaveInvitationCanceller(cancel *CancelInvitationUseCase, commandFactory ParticipationCommandFactory) *LeaveInvitationCanceller {
	return &LeaveInvitationCanceller{cancel: cancel, commandFactory: commandFactory}
}

// CancelInvitation cancels the member's invitation to the race, with the side effects of
// cancelling an invitation (docs/design/api.md).
func (canceller *LeaveInvitationCanceller) CancelInvitation(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) error {
	return canceller.cancel.Execute(ctx, canceller.commandFactory.Invitation(raceID, memberID))
}
