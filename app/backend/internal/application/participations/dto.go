package participations

import (
	"slices"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// InvitationDTO is a roster member as GET /races/{race_id}/invitations lists them. It is
// created by ParticipationDTOFactory.
type InvitationDTO struct {
	memberID   shared.MemberID
	name       string
	invited    bool
	attendance *participation.AttendanceStatus
}

// MemberID returns the roster member's ID.
func (dto InvitationDTO) MemberID() shared.MemberID { return dto.memberID }

// Name returns the roster member's name.
func (dto InvitationDTO) Name() string { return dto.name }

// Invited reports whether the member is invited to the race.
func (dto InvitationDTO) Invited() bool { return dto.invited }

// Attendance returns the member's answer, and false when they have not answered or the
// caller may not see it.
func (dto InvitationDTO) Attendance() (participation.AttendanceStatus, bool) {
	if dto.attendance == nil {
		return "", false
	}
	return *dto.attendance, true
}

// AttendanceDTO is the caller's answer as PUT /races/{race_id}/attendance returns it. It is
// created by ParticipationDTOFactory.
type AttendanceDTO struct {
	status participation.AttendanceStatus
}

// Status returns the answer.
func (dto AttendanceDTO) Status() participation.AttendanceStatus { return dto.status }

// EntryDTO is an entry as GET and POST /races/{race_id}/entries return it. It is created by
// ParticipationDTOFactory.
type EntryDTO struct {
	id      shared.EntryID
	name    string
	members []EntryMemberDTO
}

// ID returns the entry's ID.
func (dto EntryDTO) ID() shared.EntryID { return dto.id }

// Name returns the entry name.
func (dto EntryDTO) Name() string { return dto.name }

// Members returns the entry's members.
func (dto EntryDTO) Members() []EntryMemberDTO { return slices.Clone(dto.members) }

// EntryMemberDTO is a member of an entry.
type EntryMemberDTO struct {
	memberID shared.MemberID
	name     string
}

// MemberID returns the roster member's ID.
func (dto EntryMemberDTO) MemberID() shared.MemberID { return dto.memberID }

// Name returns the roster member's name.
func (dto EntryMemberDTO) Name() string { return dto.name }
