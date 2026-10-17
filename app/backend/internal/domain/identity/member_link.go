package identity

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LinkStatus is the state of a link between an account and a roster member.
type LinkStatus string

// Link states.
const (
	LinkPending  LinkStatus = "pending"
	LinkApproved LinkStatus = "approved"
	LinkRejected LinkStatus = "rejected"
	LinkUnlinked LinkStatus = "unlinked"
)

// MemberLink links an account to a roster member. An organizer approves or rejects it,
// and can unlink it after approving (ADR-0006, ADR-0008). It is created by MemberLinkFactory.
type MemberLink struct {
	id           shared.MemberLinkID
	accountID    shared.AccountID
	memberID     shared.MemberID
	inviteLinkID shared.InviteLinkID
	status       LinkStatus
	createdAt    time.Time
	decidedAt    *time.Time
	unlinkedAt   *time.Time
}

// ID returns the link's ID.
func (l *MemberLink) ID() shared.MemberLinkID { return l.id }

// AccountID returns the account.
func (l *MemberLink) AccountID() shared.AccountID { return l.accountID }

// MemberID returns the roster member.
func (l *MemberLink) MemberID() shared.MemberID { return l.memberID }

// InviteLinkID returns the invite link that created this link.
func (l *MemberLink) InviteLinkID() shared.InviteLinkID { return l.inviteLinkID }

// Status returns the state of the link.
func (l *MemberLink) Status() LinkStatus { return l.status }

// CreatedAt returns when the person logged in through the invite link.
func (l *MemberLink) CreatedAt() time.Time { return l.createdAt }

// DecidedAt returns when an organizer approved or rejected the link.
func (l *MemberLink) DecidedAt() (time.Time, bool) { return deref(l.decidedAt) }

// UnlinkedAt returns when the link was unlinked.
func (l *MemberLink) UnlinkedAt() (time.Time, bool) { return deref(l.unlinkedAt) }

// Approve approves a pending link.
func (l *MemberLink) Approve(now time.Time) error {
	return l.decide(LinkApproved, now)
}

// Reject rejects a pending link. The organizer tells the person outside the app.
func (l *MemberLink) Reject(now time.Time) error {
	return l.decide(LinkRejected, now)
}

func (l *MemberLink) decide(to LinkStatus, now time.Time) error {
	if l.status != LinkPending {
		return shared.NewStateError("the link is not waiting for approval")
	}
	l.status = to
	l.decidedAt = &now
	return nil
}

// Unlink removes an approved link, for example when the person changes their LINE account.
func (l *MemberLink) Unlink(now time.Time) error {
	if l.status != LinkApproved {
		return shared.NewStateError("the link is not approved")
	}
	l.status = LinkUnlinked
	l.unlinkedAt = &now
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
		status:       LinkPending,
		createdAt:    now,
	}, nil
}

// Reconstruct returns a link read from storage.
func (f MemberLinkFactory) Reconstruct(id shared.MemberLinkID, accountID shared.AccountID, memberID shared.MemberID, inviteLinkID shared.InviteLinkID, status LinkStatus, createdAt time.Time, decidedAt, unlinkedAt *time.Time) *MemberLink {
	return &MemberLink{id: id, accountID: accountID, memberID: memberID, inviteLinkID: inviteLinkID, status: status, createdAt: createdAt, decidedAt: decidedAt, unlinkedAt: unlinkedAt}
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
// approved, otherwise pending if one is pending, otherwise rejected if the latest link was
// rejected, otherwise none.
func (LinkStatusService) CallerStatus(links []*MemberLink) CallerStatus {
	var pending bool
	var latest *MemberLink
	for _, l := range links {
		switch l.status {
		case LinkApproved:
			return CallerApproved
		case LinkPending:
			pending = true
		}
		if latest == nil || l.createdAt.After(latest.createdAt) {
			latest = l
		}
	}
	switch {
	case pending:
		return CallerPending
	case latest != nil && latest.status == LinkRejected:
		return CallerRejected
	}
	return CallerNone
}

// RosterLinkStatus returns the link state of a roster member from all of their links, for
// roster.Member.CanIssueInviteLink and the roster list. A rejected link counts as none.
func (LinkStatusService) RosterLinkStatus(links []*MemberLink) roster.LinkStatus {
	status := roster.LinkNone
	for _, l := range links {
		switch l.status {
		case LinkApproved:
			return roster.LinkApproved
		case LinkPending:
			status = roster.LinkPending
		}
	}
	return status
}
