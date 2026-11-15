package rosters

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RosterDTOFactory assembles the DTOs of this package.
type RosterDTOFactory struct{}

// CreateRosterMember returns a roster member of the list. expiresAt is nil when the member
// has no invite link that can still be used.
func (RosterDTOFactory) CreateRosterMember(memberID shared.MemberID, name string, left bool, linkStatus roster.LinkStatus, expiresAt *time.Time) RosterMemberDTO {
	return RosterMemberDTO{memberID: memberID, name: name, left: left, linkStatus: linkStatus, inviteLinkExpiresAt: expiresAt}
}

// CreateRosterMemberDetail returns the member of the record.
func (RosterDTOFactory) CreateRosterMemberDetail(record RosterMemberDetailRecord) RosterMemberDetailDTO {
	byRace := make([]RaceLapsDTO, 0, len(record.LapsByRace))
	for _, race := range record.LapsByRace {
		laps := make([]LapDTO, 0, len(race.Laps))
		for _, lap := range race.Laps {
			laps = append(laps, LapDTO{id: lap.ID, lapTime: lap.Time})
		}
		byRace = append(byRace, RaceLapsDTO{raceID: race.RaceID, raceName: race.RaceName, laps: laps})
	}
	return RosterMemberDetailDTO{memberID: record.MemberID, name: record.Name, left: record.Left, average: record.Average, lapCount: record.LapCount, lapsByRace: byRace}
}

// CreateInviteLink returns a new invite link.
func (RosterDTOFactory) CreateInviteLink(token string, expiresAt time.Time) InviteLinkDTO {
	return InviteLinkDTO{token: token, expiresAt: expiresAt}
}

// CreatePendingLink returns a link that waits for approval.
func (RosterDTOFactory) CreatePendingLink(record PendingLinkRecord) PendingLinkDTO {
	return PendingLinkDTO{linkID: record.LinkID, lineDisplayName: record.LineDisplayName, memberID: record.MemberID, memberName: record.MemberName, createdAt: record.CreatedAt}
}
