package auth

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// MeDTO is the caller as GET /me returns it. It is created by MeDTOFactory.
type MeDTO struct {
	displayName  string
	isOrganizer  bool
	status       identity.CallerStatus
	rosterMember *RosterMemberDTO
}

// DisplayName returns the LINE display name.
func (dto MeDTO) DisplayName() string { return dto.displayName }

// IsOrganizer reports whether the caller is an organizer (ADR-0006).
func (dto MeDTO) IsOrganizer() bool { return dto.isOrganizer }

// Status returns the caller's link status.
func (dto MeDTO) Status() identity.CallerStatus { return dto.status }

// RosterMember returns the linked roster member, set only when the status is approved.
func (dto MeDTO) RosterMember() (RosterMemberDTO, bool) {
	if dto.rosterMember == nil {
		return RosterMemberDTO{}, false
	}
	return *dto.rosterMember, true
}

// RosterMemberDTO is the roster member a user acts as.
type RosterMemberDTO struct {
	id   shared.MemberID
	name string
}

// ID returns the roster member's ID.
func (dto RosterMemberDTO) ID() shared.MemberID { return dto.id }

// Name returns the roster member's name.
func (dto RosterMemberDTO) Name() string { return dto.name }

// SessionDTO is a new session. The token goes into the cookie and is never stored.
// It is created by SessionDTOFactory.
type SessionDTO struct {
	token     string
	expiresAt time.Time
}

// Token returns the session token.
func (dto SessionDTO) Token() string { return dto.token }

// ExpiresAt returns when the session ends.
func (dto SessionDTO) ExpiresAt() time.Time { return dto.expiresAt }
