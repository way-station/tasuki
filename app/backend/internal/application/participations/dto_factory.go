package participations

import (
	"slices"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ParticipationDTOFactory assembles the DTOs of this package.
type ParticipationDTOFactory struct{}

// CreateInvitation returns a roster member of the invitation list. The answer is left out
// when canSeeAnswers is false (participation.AttendancePolicy).
func (ParticipationDTOFactory) CreateInvitation(memberID shared.MemberID, name string, invited bool, attendance *participation.AttendanceStatus, canSeeAnswers bool) InvitationDTO {
	if !canSeeAnswers {
		attendance = nil
	}
	return InvitationDTO{memberID: memberID, name: name, invited: invited, attendance: attendance}
}

// CreateAttendance returns an answer.
func (ParticipationDTOFactory) CreateAttendance(status participation.AttendanceStatus) AttendanceDTO {
	return AttendanceDTO{status: status}
}

// CreateEntry returns an entry with its members.
func (ParticipationDTOFactory) CreateEntry(id shared.EntryID, name string, members []EntryMemberDTO) EntryDTO {
	return EntryDTO{id: id, name: name, members: slices.Clone(members)}
}

// CreateEntryMember returns a member of an entry.
func (ParticipationDTOFactory) CreateEntryMember(memberID shared.MemberID, name string) EntryMemberDTO {
	return EntryMemberDTO{memberID: memberID, name: name}
}
