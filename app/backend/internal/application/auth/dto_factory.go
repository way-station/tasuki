package auth

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// MeDTOFactory assembles MeDTO values.
type MeDTOFactory struct{}

// Create returns the caller. rosterMember is ignored unless status is approved.
func (MeDTOFactory) Create(displayName string, isOrganizer bool, status identity.CallerStatus, rosterMember *RosterMemberDTO) MeDTO {
	if status != identity.CallerApproved {
		rosterMember = nil
	}
	return MeDTO{displayName: displayName, isOrganizer: isOrganizer, status: status, rosterMember: rosterMember}
}

// CreateRosterMember returns a roster member.
func (MeDTOFactory) CreateRosterMember(id shared.MemberID, name string) *RosterMemberDTO {
	return &RosterMemberDTO{id: id, name: name}
}

// SessionDTOFactory assembles SessionDTO values.
type SessionDTOFactory struct{}

// Create returns a new session.
func (SessionDTOFactory) Create(token string, expiresAt time.Time) SessionDTO {
	return SessionDTO{token: token, expiresAt: expiresAt}
}
