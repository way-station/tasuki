// Package rosters holds the use cases of the team roster and the approval of member links
// (docs/design/api.md, ADR-0011).
package rosters

import "github.com/way-station/tasuki/app/backend/internal/domain/shared"

// AddRosterMemberCommand is a roster member to add. It is created by RosterCommandFactory.
type AddRosterMemberCommand struct {
	name string
}

// SetLeftCommand marks a roster member as left, or undoes it. It is created by
// RosterCommandFactory.
type SetLeftCommand struct {
	memberID shared.MemberID
	left     bool
}

// RosterCommandFactory assembles the commands of this package. A blank name is left to the
// domain.
type RosterCommandFactory struct{}

// AddRosterMember returns the command to add a roster member named name.
func (RosterCommandFactory) AddRosterMember(name string) AddRosterMemberCommand {
	return AddRosterMemberCommand{name: name}
}

// SetLeft returns the command to mark the member as left, or to undo it.
func (RosterCommandFactory) SetLeft(memberID shared.MemberID, left bool) SetLeftCommand {
	return SetLeftCommand{memberID: memberID, left: left}
}
