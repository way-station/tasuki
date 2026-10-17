package identity

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// MemberLink links an account to a roster member. An organizer approves or rejects it,
// and can unlink it after approving (ADR-0006, ADR-0008). Its states are booleans, as in the
// member_links table: approved is nil while pending, true when approved, false when rejected.
// It is created by MemberLinkFactory.
type MemberLink struct {
	id           shared.MemberLinkID
	accountID    shared.AccountID
	memberID     shared.MemberID
	inviteLinkID shared.InviteLinkID
	approved     *bool
	unlinked     bool
	createdAt    time.Time
	decidedAt    *time.Time
}

// ID returns the link's ID.
func (l *MemberLink) ID() shared.MemberLinkID { return l.id }

// AccountID returns the account.
func (l *MemberLink) AccountID() shared.AccountID { return l.accountID }

// MemberID returns the roster member.
func (l *MemberLink) MemberID() shared.MemberID { return l.memberID }

// InviteLinkID returns the invite link that created this link.
func (l *MemberLink) InviteLinkID() shared.InviteLinkID { return l.inviteLinkID }

// CreatedAt returns when the person logged in through the invite link.
func (l *MemberLink) CreatedAt() time.Time { return l.createdAt }

// DecidedAt returns when an organizer approved or rejected the link.
func (l *MemberLink) DecidedAt() (time.Time, bool) { return deref(l.decidedAt) }

// Unlinked reports whether an organizer unlinked the approved link.
func (l *MemberLink) Unlinked() bool { return l.unlinked }

// Active reports whether the link is approved and not unlinked: the account acts as the
// roster member.
func (l *MemberLink) Active() bool { return l.approved != nil && *l.approved && !l.unlinked }

// Approved returns whether an organizer approved (true) or rejected (false) the link, and
// false as the second value while it is pending.
func (l *MemberLink) Approved() (approved, decided bool) {
	if l.approved == nil {
		return false, false
	}
	return *l.approved, true
}

// Pending reports whether the link is waiting for approval.
func (l *MemberLink) Pending() bool { return l.approved == nil }

// Rejected reports whether an organizer rejected the link.
func (l *MemberLink) Rejected() bool { return l.approved != nil && !*l.approved }

// Approve approves a pending link.
func (l *MemberLink) Approve(now time.Time) error {
	return l.decide(true, now)
}

// Reject rejects a pending link. The organizer tells the person outside the app.
func (l *MemberLink) Reject(now time.Time) error {
	return l.decide(false, now)
}

func (l *MemberLink) decide(approved bool, now time.Time) error {
	if !l.Pending() {
		return shared.NewStateError("the link is not waiting for approval")
	}
	l.approved = &approved
	l.decidedAt = &now
	return nil
}

// Unlink removes an approved link, for example when the person changes their LINE account.
// The link stays approved, with Unlinked set.
func (l *MemberLink) Unlink() error {
	if !l.Active() {
		return shared.NewStateError("the link is not approved")
	}
	l.unlinked = true
	return nil
}

// MemberLinkFactory assembles MemberLink values.
type MemberLinkFactory struct {
	ids shared.IDGenerator
}

// NewMemberLinkFactory returns a MemberLinkFactory.
func NewMemberLinkFactory(ids shared.IDGenerator) MemberLinkFactory {
	return MemberLinkFactory{ids: ids}
}

// New returns a pending link between the account and the invite link's roster member, and
// marks the invite link as used. It returns ErrInviteLinkUnusable when the invite link
// cannot be used.
func (f MemberLinkFactory) New(accountID shared.AccountID, invite *InviteLink, now time.Time) (*MemberLink, error) {
	if err := invite.use(now); err != nil {
		return nil, err
	}
	return &MemberLink{
		id:           shared.MemberLinkID(f.ids.NewID()),
		accountID:    accountID,
		memberID:     invite.memberID,
		inviteLinkID: invite.id,
		createdAt:    now,
	}, nil
}

// Reconstruct returns a link read from storage. approved is nil for a pending link.
func (f MemberLinkFactory) Reconstruct(id shared.MemberLinkID, accountID shared.AccountID, memberID shared.MemberID, inviteLinkID shared.InviteLinkID, approved *bool, unlinked bool, createdAt time.Time, decidedAt *time.Time) *MemberLink {
	return &MemberLink{id: id, accountID: accountID, memberID: memberID, inviteLinkID: inviteLinkID, approved: approved, unlinked: unlinked, createdAt: createdAt, decidedAt: decidedAt}
}

// CallerStatus is what GET /me reports about an account's link (docs/design/api.md).
type CallerStatus string

// Caller states.
const (
	CallerApproved CallerStatus = "approved"
	CallerPending  CallerStatus = "pending"
	CallerRejected CallerStatus = "rejected"
	CallerNone     CallerStatus = "none"
)

// LinkStatusService derives statuses from an account's or a roster member's links.
type LinkStatusService struct{}

// CallerStatus returns the status of an account from all of its links: approved if one is
// approved and not unlinked, otherwise pending if one is pending, otherwise rejected if the
// latest link was rejected, otherwise none.
func (LinkStatusService) CallerStatus(links []*MemberLink) CallerStatus {
	var pending bool
	var latest *MemberLink
	for _, l := range links {
		switch {
		case l.Active():
			return CallerApproved
		case l.Pending():
			pending = true
		}
		if latest == nil || l.createdAt.After(latest.createdAt) {
			latest = l
		}
	}
	switch {
	case pending:
		return CallerPending
	case latest != nil && latest.Rejected():
		return CallerRejected
	}
	return CallerNone
}

// RosterLinkStatus returns the link state of a roster member from all of their links, for
// roster.Member.CanIssueInviteLink and the roster list. Rejected and unlinked links count as
// none.
func (LinkStatusService) RosterLinkStatus(links []*MemberLink) roster.LinkStatus {
	status := roster.LinkNone
	for _, l := range links {
		switch {
		case l.Active():
			return roster.LinkApproved
		case l.Pending():
			status = roster.LinkPending
		}
	}
	return status
}
