package identity

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// MemberLink links a user to a roster member. An organizer approves or rejects it,
// and can unlink it after approving (ADR-0006, ADR-0008). Its states are timestamps, as in
// the member_links table: a state holds when its time is set, and a link with none of them
// set is pending. It is created by MemberLinkFactory.
type MemberLink struct {
	id           shared.MemberLinkID
	userID       shared.UserID
	memberID     shared.MemberID
	inviteLinkID shared.InviteLinkID
	createdAt    time.Time
	approvedAt   *time.Time
	rejectedAt   *time.Time
	unlinkedAt   *time.Time // set only after approvedAt; approvedAt stays
}

// ID returns the link's ID.
func (link *MemberLink) ID() shared.MemberLinkID { return link.id }

// UserID returns the user.
func (link *MemberLink) UserID() shared.UserID { return link.userID }

// MemberID returns the roster member.
func (link *MemberLink) MemberID() shared.MemberID { return link.memberID }

// InviteLinkID returns the invite link that created this link.
func (link *MemberLink) InviteLinkID() shared.InviteLinkID { return link.inviteLinkID }

// CreatedAt returns when the person logged in through the invite link.
func (link *MemberLink) CreatedAt() time.Time { return link.createdAt }

// ApprovedAt returns when an organizer approved the link, and false if they have not.
func (link *MemberLink) ApprovedAt() (time.Time, bool) { return deref(link.approvedAt) }

// RejectedAt returns when an organizer rejected the link, and false if they have not.
func (link *MemberLink) RejectedAt() (time.Time, bool) { return deref(link.rejectedAt) }

// UnlinkedAt returns when an organizer unlinked the link, and false if they have not.
func (link *MemberLink) UnlinkedAt() (time.Time, bool) { return deref(link.unlinkedAt) }

// Pending reports whether the link is waiting for approval: neither approved nor rejected.
func (link *MemberLink) Pending() bool { return link.approvedAt == nil && link.rejectedAt == nil }

// Rejected reports whether an organizer rejected the link.
func (link *MemberLink) Rejected() bool { return link.rejectedAt != nil }

// Unlinked reports whether an organizer unlinked the approved link.
func (link *MemberLink) Unlinked() bool { return link.unlinkedAt != nil }

// Active reports whether the link is approved and not unlinked: the user acts as the
// roster member.
func (link *MemberLink) Active() bool { return link.approvedAt != nil && link.unlinkedAt == nil }

// Approve approves a pending link.
func (link *MemberLink) Approve(now time.Time) error {
	if !link.Pending() {
		return shared.NewStateError("the link is not waiting for approval")
	}
	link.approvedAt = &now
	return nil
}

// Reject rejects a pending link. The organizer tells the person outside the app.
func (link *MemberLink) Reject(now time.Time) error {
	if !link.Pending() {
		return shared.NewStateError("the link is not waiting for approval")
	}
	link.rejectedAt = &now
	return nil
}

// Unlink removes an approved link, for example when the person changes their LINE account.
// The approval time stays.
func (link *MemberLink) Unlink(now time.Time) error {
	if !link.Active() {
		return shared.NewStateError("the link is not approved")
	}
	link.unlinkedAt = &now
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

// New returns a pending link between the user and the invite link's roster member, and
// marks the invite link as used. It returns ErrInviteLinkUnusable when the invite link
// cannot be used.
func (factory MemberLinkFactory) New(userID shared.UserID, invite *InviteLink, now time.Time) (*MemberLink, error) {
	if err := invite.use(now); err != nil {
		return nil, err
	}
	return &MemberLink{
		id:           shared.MemberLinkID(factory.ids.NewID()),
		userID:       userID,
		memberID:     invite.memberID,
		inviteLinkID: invite.id,
		createdAt:    now,
	}, nil
}

// Reconstruct returns a link read from storage. A nil time means the state does not hold.
func (factory MemberLinkFactory) Reconstruct(id shared.MemberLinkID, userID shared.UserID, memberID shared.MemberID, inviteLinkID shared.InviteLinkID, createdAt time.Time, approvedAt, rejectedAt, unlinkedAt *time.Time) *MemberLink {
	return &MemberLink{id: id, userID: userID, memberID: memberID, inviteLinkID: inviteLinkID, createdAt: createdAt, approvedAt: approvedAt, rejectedAt: rejectedAt, unlinkedAt: unlinkedAt}
}

// CallerStatus is what GET /me reports about a user's link (docs/design/api.md).
type CallerStatus string

// Caller states.
const (
	CallerApproved CallerStatus = "approved"
	CallerPending  CallerStatus = "pending"
	CallerRejected CallerStatus = "rejected"
	CallerNone     CallerStatus = "none"
)

// LinkStatusService derives statuses from a user's or a roster member's links.
type LinkStatusService struct{}

// CallerStatus returns the status of a user from all of its links: approved if one is
// approved and not unlinked, otherwise pending if one is pending, otherwise rejected if the
// latest link was rejected, otherwise none.
func (LinkStatusService) CallerStatus(links []*MemberLink) CallerStatus {
	var pending bool
	var latest *MemberLink
	for _, link := range links {
		switch {
		case link.Active():
			return CallerApproved
		case link.Pending():
			pending = true
		}
		if latest == nil || link.createdAt.After(latest.createdAt) {
			latest = link
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
	for _, link := range links {
		switch {
		case link.Active():
			return roster.LinkApproved
		case link.Pending():
			status = roster.LinkPending
		}
	}
	return status
}
