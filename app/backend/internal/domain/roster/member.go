// Package roster is the domain of the team roster: the roster members and the rules
// for marking them as left and for issuing invite links (docs/design/database.md).
package roster

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Member is a person on the team roster. It is created by MemberFactory.
type Member struct {
	id       shared.MemberID
	teamID   shared.TeamID
	name     string
	leftAt   *time.Time
	average  shared.Pace // per km; DefaultAverage until the first lap
	lapCount int         // laps counted in the average
}

// ID returns the member's ID.
func (member *Member) ID() shared.MemberID { return member.id }

// TeamID returns the team the member belongs to.
func (member *Member) TeamID() shared.TeamID { return member.teamID }

// Name returns the member's name. Names cannot be changed.
func (member *Member) Name() string { return member.name }

// Average returns the average time per km. AverageService keeps it up to date.
func (member *Member) Average() shared.Pace { return member.average }

// LapCount returns the number of laps counted in the average. The default average is not
// counted.
func (member *Member) LapCount() int { return member.lapCount }

// Left reports whether the member has left.
func (member *Member) Left() bool { return member.leftAt != nil }

// LeftAt returns when the member left, and false if they have not left.
func (member *Member) LeftAt() (time.Time, bool) {
	if member.leftAt == nil {
		return time.Time{}, false
	}
	return *member.leftAt, true
}

// Leave marks the member as left. Marking a left member again does nothing.
func (member *Member) Leave(now time.Time) {
	if member.leftAt != nil {
		return
	}
	member.leftAt = &now
}

// Rejoin undoes Leave. Cancelled invitations do not come back.
func (member *Member) Rejoin() {
	member.leftAt = nil
}

// LinkStatus is the state of the link between a roster member and an account,
// as far as issuing invite links is concerned.
type LinkStatus int

// Link states.
const (
	LinkNone LinkStatus = iota
	LinkPending
	LinkApproved
)

// CanIssueInviteLink returns a StateError when an invite link may not be created for the
// member: they have left, or they already have an approved or pending link.
func (member *Member) CanIssueInviteLink(link LinkStatus) error {
	switch {
	case member.Left():
		return shared.NewStateError("the member has left")
	case link == LinkApproved:
		return shared.NewStateError("the member already has an approved account")
	case link == LinkPending:
		return shared.NewStateError("the member has a link waiting for approval")
	}
	return nil
}
