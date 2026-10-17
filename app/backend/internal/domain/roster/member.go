// Package roster is the domain of the team roster: the roster members and the rules
// for marking them as left and for issuing invite links (docs/design/database.md).
package roster

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Member is a person on the team roster. It is created by MemberFactory.
type Member struct {
	id     shared.MemberID
	teamID shared.TeamID
	name   string
	leftAt *time.Time
}

// ID returns the member's ID.
func (m *Member) ID() shared.MemberID { return m.id }

// TeamID returns the team the member belongs to.
func (m *Member) TeamID() shared.TeamID { return m.teamID }

// Name returns the member's name. Names cannot be changed.
func (m *Member) Name() string { return m.name }

// Left reports whether the member has left.
func (m *Member) Left() bool { return m.leftAt != nil }

// LeftAt returns when the member left, and false if they have not left.
func (m *Member) LeftAt() (time.Time, bool) {
	if m.leftAt == nil {
		return time.Time{}, false
	}
	return *m.leftAt, true
}

// Leave marks the member as left. Marking a left member again does nothing.
func (m *Member) Leave(now time.Time) {
	if m.leftAt != nil {
		return
	}
	m.leftAt = &now
}

// Rejoin undoes Leave. Cancelled invitations do not come back.
func (m *Member) Rejoin() {
	m.leftAt = nil
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
func (m *Member) CanIssueInviteLink(link LinkStatus) error {
	switch {
	case m.Left():
		return shared.NewStateError("the member has left")
	case link == LinkApproved:
		return shared.NewStateError("the member already has an approved account")
	case link == LinkPending:
		return shared.NewStateError("the member has a link waiting for approval")
	}
	return nil
}
