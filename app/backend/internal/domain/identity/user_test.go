package identity

import (
	"bytes"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type seqIDs struct{ n int }

func (ids *seqIDs) NewID() string { ids.n++; return "id" + strconv.Itoa(ids.n) }

type fixedToken string

func (token fixedToken) NewToken() (string, error) { return string(token), nil }

var now = time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

func newInvite(t *testing.T) (*InviteLink, string) {
	t.Helper()
	inviteLink, token, err := NewInviteLinkFactory(&seqIDs{}, fixedToken("secret")).New("m1", now)
	if err != nil {
		t.Fatal(err)
	}
	return inviteLink, token
}

func TestOrganizerPolicy(t *testing.T) {
	factory := NewUserFactory(&seqIDs{})
	org, _ := factory.New("U-org", "Org")
	member, _ := factory.New("U-member", "Member")
	policy := NewOrganizerPolicy([]string{"U-org"})
	if !policy.IsOrganizer(org) || policy.IsOrganizer(member) {
		t.Error("IsOrganizer() is wrong")
	}
	if _, err := factory.New(" ", "x"); err == nil {
		t.Error("New(blank LINE user ID): want an error")
	}
}

func TestInviteLink(t *testing.T) {
	inviteLink, token := newInvite(t)
	if token != "secret" || !bytes.Equal(inviteLink.TokenHash(), HashToken("secret")) {
		t.Error("the token hash does not match the token")
	}
	if !inviteLink.Usable(now) || !inviteLink.Usable(now.Add(InviteLinkLifetime-time.Second)) {
		t.Error("a new link must be usable for one week")
	}
	if inviteLink.Usable(now.Add(InviteLinkLifetime)) {
		t.Error("a link must not be usable after one week")
	}

	inviteLink.Revoke(now)
	if inviteLink.Usable(now) {
		t.Error("a replaced link must not be usable")
	}
}

func TestMemberLinkFactoryUsesTheInviteOnce(t *testing.T) {
	factory := NewMemberLinkFactory(&seqIDs{})
	invite, _ := newInvite(t)

	link, err := factory.New("a1", invite, now)
	if err != nil {
		t.Fatal(err)
	}
	if !link.Pending() || link.MemberID() != "m1" || link.InviteLinkID() != invite.ID() {
		t.Errorf("New() = %+v", link)
	}
	if _, err := factory.New("a2", invite, now); !errors.Is(err, ErrInviteLinkUnusable) {
		t.Errorf("second use: error = %v, want ErrInviteLinkUnusable", err)
	}
}

func TestMemberLinkTransitions(t *testing.T) {
	isState := func(err error) bool { var se *shared.StateError; return errors.As(err, &se) }
	newLink := func() *MemberLink {
		invite, _ := newInvite(t)
		memberLink, err := NewMemberLinkFactory(&seqIDs{}).New("a1", invite, now)
		if err != nil {
			t.Fatal(err)
		}
		return memberLink
	}

	memberLink := newLink()
	if !isState(memberLink.Unlink(now)) {
		t.Error("Unlink() of a pending link: want a StateError")
	}
	if err := memberLink.Approve(now); err != nil || !memberLink.Active() || memberLink.Pending() {
		t.Fatalf("Approve() = %v, active %v", err, memberLink.Active())
	}
	if at, ok := memberLink.ApprovedAt(); !ok || !at.Equal(now) {
		t.Errorf("ApprovedAt() = %v, %v, want %v", at, ok, now)
	}
	if !isState(memberLink.Reject(now)) {
		t.Error("Reject() of an approved link: want a StateError")
	}
	if err := memberLink.Unlink(now); err != nil || !memberLink.Unlinked() || memberLink.Active() {
		t.Errorf("Unlink() = %v, unlinked %v, active %v", err, memberLink.Unlinked(), memberLink.Active())
	}
	if _, ok := memberLink.ApprovedAt(); !ok {
		t.Error("an unlinked link keeps its approval time")
	}
	if !isState(memberLink.Unlink(now)) {
		t.Error("Unlink() twice: want a StateError")
	}

	rejectedLink := newLink()
	if err := rejectedLink.Reject(now); err != nil || !rejectedLink.Rejected() || rejectedLink.Pending() {
		t.Errorf("Reject() = %v, rejected %v", err, rejectedLink.Rejected())
	}
	if !isState(rejectedLink.Unlink(now)) {
		t.Error("Unlink() of a rejected link: want a StateError")
	}
}

func TestLinkStatusService(t *testing.T) {
	factory := NewMemberLinkFactory(&seqIDs{})
	pendingLink := func(at time.Time) *MemberLink { return factory.Reconstruct("l", "a1", "m1", "i", at, nil, nil, nil) }
	approvedLink := func(at time.Time) *MemberLink { return factory.Reconstruct("l", "a1", "m1", "i", at, &at, nil, nil) }
	rejectedLink := func(at time.Time) *MemberLink { return factory.Reconstruct("l", "a1", "m1", "i", at, nil, &at, nil) }
	unlinked := func(at time.Time) *MemberLink { return factory.Reconstruct("l", "a1", "m1", "i", at, &at, nil, &at) }
	earlier, later := now, now.Add(time.Hour)
	svc := LinkStatusService{}

	callerTests := []struct {
		name  string
		links []*MemberLink
		want  CallerStatus
	}{
		{name: "no links", want: CallerNone},
		{name: "approved", links: []*MemberLink{rejectedLink(earlier), approvedLink(later)}, want: CallerApproved},
		{name: "pending", links: []*MemberLink{pendingLink(later)}, want: CallerPending},
		{name: "approved after pending", links: []*MemberLink{pendingLink(earlier), approvedLink(later)}, want: CallerApproved},
		{name: "pending before a later rejection", links: []*MemberLink{pendingLink(earlier), rejectedLink(later)}, want: CallerPending},
		{name: "latest rejected", links: []*MemberLink{unlinked(earlier), rejectedLink(later)}, want: CallerRejected},
		{name: "latest unlinked", links: []*MemberLink{rejectedLink(earlier), unlinked(later)}, want: CallerNone},
	}
	for _, testCase := range callerTests {
		t.Run(testCase.name, func(t *testing.T) {
			if got := svc.CallerStatus(testCase.links); got != testCase.want {
				t.Errorf("CallerStatus() = %s, want %s", got, testCase.want)
			}
		})
	}

	if got := svc.RosterLinkStatus([]*MemberLink{rejectedLink(earlier)}); got != roster.LinkNone {
		t.Errorf("RosterLinkStatus(rejected) = %v, want LinkNone", got)
	}
	if got := svc.RosterLinkStatus([]*MemberLink{unlinked(earlier)}); got != roster.LinkNone {
		t.Errorf("RosterLinkStatus(unlinked) = %v, want LinkNone", got)
	}
	if got := svc.RosterLinkStatus([]*MemberLink{pendingLink(earlier)}); got != roster.LinkPending {
		t.Errorf("RosterLinkStatus(pending) = %v, want LinkPending", got)
	}
}
