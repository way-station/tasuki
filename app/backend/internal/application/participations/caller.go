package participations

import (
	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// rosterMemberOf returns the roster member the caller acts as, and false when the caller
// is not linked, such as an organizer without a link.
func rosterMemberOf(caller auth.MeDTO) (shared.MemberID, bool) {
	rosterMember, ok := caller.RosterMember()
	if !ok {
		return "", false
	}
	return rosterMember.ID(), true
}
