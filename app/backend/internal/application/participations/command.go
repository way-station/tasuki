// Package participations holds the use cases of who takes part in a race: invitations,
// attendance answers, and entries (docs/design/api.md, ADR-0011).
package participations

import (
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// InvitationCommand is a roster member to invite to a race, or whose invitation to cancel.
// It is created by ParticipationCommandFactory.
type InvitationCommand struct {
	raceID   shared.RaceID
	memberID shared.MemberID
}

// AttendanceCommand is the caller's answer for a race. It is created by
// ParticipationCommandFactory.
type AttendanceCommand struct {
	raceID shared.RaceID
	status participation.AttendanceStatus
}

// CreateEntryCommand is an entry to create in a race. It is created by
// ParticipationCommandFactory.
type CreateEntryCommand struct {
	raceID shared.RaceID
	name   string
}

// EntryMemberCommand is a roster member to add to an entry, or to remove from it. It is
// created by ParticipationCommandFactory.
type EntryMemberCommand struct {
	raceID   shared.RaceID
	entryID  shared.EntryID
	memberID shared.MemberID
}

// ParticipationCommandFactory assembles the commands of this package. The rules of the
// values, such as the attendance answers and a blank entry name, are left to the domain.
type ParticipationCommandFactory struct{}

// Invitation returns the command to invite the member to the race, or to cancel it.
func (ParticipationCommandFactory) Invitation(raceID shared.RaceID, memberID shared.MemberID) InvitationCommand {
	return InvitationCommand{raceID: raceID, memberID: memberID}
}

// Attendance returns the command to answer status for the race.
func (ParticipationCommandFactory) Attendance(raceID shared.RaceID, status string) AttendanceCommand {
	return AttendanceCommand{raceID: raceID, status: participation.AttendanceStatus(status)}
}

// CreateEntry returns the command to create an entry named name in the race.
func (ParticipationCommandFactory) CreateEntry(raceID shared.RaceID, name string) CreateEntryCommand {
	return CreateEntryCommand{raceID: raceID, name: name}
}

// EntryMember returns the command to add the member to the entry of the race, or to remove
// them from it.
func (ParticipationCommandFactory) EntryMember(raceID shared.RaceID, entryID shared.EntryID, memberID shared.MemberID) EntryMemberCommand {
	return EntryMemberCommand{raceID: raceID, entryID: entryID, memberID: memberID}
}
