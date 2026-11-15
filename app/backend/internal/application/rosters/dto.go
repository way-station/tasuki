package rosters

import (
	"slices"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RosterMemberDTO is a roster member as the roster list returns them. It is created by
// RosterDTOFactory.
type RosterMemberDTO struct {
	memberID            shared.MemberID
	name                string
	left                bool
	linkStatus          roster.LinkStatus
	inviteLinkExpiresAt *time.Time
}

// MemberID returns the roster member's ID.
func (dto RosterMemberDTO) MemberID() shared.MemberID { return dto.memberID }

// Name returns the roster member's name.
func (dto RosterMemberDTO) Name() string { return dto.name }

// Left reports whether the member has left.
func (dto RosterMemberDTO) Left() bool { return dto.left }

// LinkStatus returns the state of the member's link. A rejected or unlinked link is none.
func (dto RosterMemberDTO) LinkStatus() roster.LinkStatus { return dto.linkStatus }

// InviteLinkExpiresAt returns when the member's invite link expires, and false when they
// have no invite link that can still be used.
func (dto RosterMemberDTO) InviteLinkExpiresAt() (time.Time, bool) {
	if dto.inviteLinkExpiresAt == nil {
		return time.Time{}, false
	}
	return *dto.inviteLinkExpiresAt, true
}

// RosterMemberDetailDTO is a roster member with their average time and laps by race. It is
// created by RosterDTOFactory.
type RosterMemberDetailDTO struct {
	memberID   shared.MemberID
	name       string
	left       bool
	average    shared.Pace
	lapCount   int
	lapsByRace []RaceLapsDTO
}

// MemberID returns the roster member's ID.
func (dto RosterMemberDetailDTO) MemberID() shared.MemberID { return dto.memberID }

// Name returns the roster member's name.
func (dto RosterMemberDetailDTO) Name() string { return dto.name }

// Left reports whether the member has left.
func (dto RosterMemberDetailDTO) Left() bool { return dto.left }

// Average returns the average time per km.
func (dto RosterMemberDetailDTO) Average() shared.Pace { return dto.average }

// LapCount returns the number of laps counted in the average.
func (dto RosterMemberDetailDTO) LapCount() int { return dto.lapCount }

// LapsByRace returns the member's laps by race.
func (dto RosterMemberDetailDTO) LapsByRace() []RaceLapsDTO { return slices.Clone(dto.lapsByRace) }

// RaceLapsDTO is the member's laps for one race.
type RaceLapsDTO struct {
	raceID   shared.RaceID
	raceName string
	laps     []LapDTO
}

// RaceID returns the race.
func (dto RaceLapsDTO) RaceID() shared.RaceID { return dto.raceID }

// RaceName returns the race name.
func (dto RaceLapsDTO) RaceName() string { return dto.raceName }

// Laps returns the laps in submission order.
func (dto RaceLapsDTO) Laps() []LapDTO { return slices.Clone(dto.laps) }

// LapDTO is one lap.
type LapDTO struct {
	id      shared.LapID
	lapTime shared.Pace
}

// ID returns the lap's ID.
func (dto LapDTO) ID() shared.LapID { return dto.id }

// Time returns the lap time per km.
func (dto LapDTO) Time() shared.Pace { return dto.lapTime }

// InviteLinkDTO is a new invite link. The token goes into the URL and is never stored. It is
// created by RosterDTOFactory.
type InviteLinkDTO struct {
	token     string
	expiresAt time.Time
}

// Token returns the token of the link.
func (dto InviteLinkDTO) Token() string { return dto.token }

// ExpiresAt returns when the link expires.
func (dto InviteLinkDTO) ExpiresAt() time.Time { return dto.expiresAt }

// PendingLinkDTO is a link that waits for approval. It is created by RosterDTOFactory.
type PendingLinkDTO struct {
	linkID          shared.MemberLinkID
	lineDisplayName string
	memberID        shared.MemberID
	memberName      string
	createdAt       time.Time
}

// LinkID returns the link's ID.
func (dto PendingLinkDTO) LinkID() shared.MemberLinkID { return dto.linkID }

// LineDisplayName returns the LINE display name of the user.
func (dto PendingLinkDTO) LineDisplayName() string { return dto.lineDisplayName }

// MemberID returns the roster member's ID.
func (dto PendingLinkDTO) MemberID() shared.MemberID { return dto.memberID }

// MemberName returns the roster member's name.
func (dto PendingLinkDTO) MemberName() string { return dto.memberName }

// CreatedAt returns when the person logged in through the invite link.
func (dto PendingLinkDTO) CreatedAt() time.Time { return dto.createdAt }
