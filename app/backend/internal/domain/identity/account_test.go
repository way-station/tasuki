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

func (s *seqIDs) NewID() string { s.n++; return "id" + strconv.Itoa(s.n) }

type fixedToken string

func (f fixedToken) NewToken() (string, error) { return string(f), nil }

var now = time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)

func newInvite(t *testing.T) (*InviteLink, string) {
	t.Helper()
	l, token, err := NewInviteLinkFactory(&seqIDs{}, fixedToken("secret")).New("m1", now)
	if err != nil {
		t.Fatal(err)
	}
	return l, token
}

func TestOrganizerPolicy(t *testing.T) {
	f := NewAccountFactory(&seqIDs{})
	org, _ := f.New("U-org", "Org")
	member, _ := f.New("U-member", "Member")
	p := NewOrganizerPolicy([]string{"U-org"})
	if !p.IsOrganizer(org) || p.IsOrganizer(member) {
		t.Error("IsOrganizer() is wrong")
	}
	if _, err := f.New(" ", "x"); err == nil {
		t.Error("New(blank LINE user ID): want an error")
	}
}

func TestInviteLink(t *testing.T) {
	l, token := newInvite(t)
	if token != "secret" || !bytes.Equal(l.TokenHash(), HashToken("secret")) {
		t.Error("the token hash does not match the token")
	}
	if !l.Usable(now) || !l.Usable(now.Add(InviteLinkLifetime-time.Second)) {
		t.Error("a new link must be usable for one week")
	}
	if l.Usable(now.Add(InviteLinkLifetime)) {
		t.Error("a link must not be usable after one week")
	}

	l.Revoke(now)
	if l.Usable(now) {
		t.Error("a replaced link must not be usable")
	}
}

func TestMemberLinkFactoryUsesTheInviteOnce(t *testing.T) {
	f := NewMemberLinkFactory(&seqIDs{})
	invite, _ := newInvite(t)

	link, err := f.New("a1", invite, now)
	if err != nil {
		t.Fatal(err)
	}
	if !link.Pending() || link.MemberID() != "m1" || link.InviteLinkID() != invite.ID() {
		t.Errorf("New() = %+v", link)
	}
	if _, err := f.New("a2", invite, now); !errors.Is(err, ErrInviteLinkUnusable) {
		t.Errorf("second use: error = %v, want ErrInviteLinkUnusable", err)
	}
}

func TestMemberLinkTransitions(t *testing.T) {
	isState := func(err error) bool { var se *shared.StateError; return errors.As(err, &se) }
	newLink := func() *MemberLink {
		invite, _ := newInvite(t)
		l, err := NewMemberLinkFactory(&seqIDs{}).New("a1", invite, now)
		if err != nil {
			t.Fatal(err)
		}
		return l
	}

	l := newLink()
	if !isState(l.Unlink()) {
		t.Error("Unlink() of a pending link: want a StateError")
	}
	if err := l.Approve(now); err != nil || !l.Active() || l.Pending() {
		t.Fatalf("Approve() = %v, active %v", err, l.Active())
	}
	if approved, decided := l.Approved(); !approved || !decided {
		t.Errorf("Approved() = %v, %v, want true, true", approved, decided)
	}
	if !isState(l.Reject(now)) {
		t.Error("Reject() of an approved link: want a StateError")
	}
	if err := l.Unlink(); err != nil || !l.Unlinked() || l.Active() {
		t.Errorf("Unlink() = %v, unlinked %v, active %v", err, l.Unlinked(), l.Active())
	}
	if approved, _ := l.Approved(); !approved {
		t.Error("an unlinked link stays approved")
	}
	if !isState(l.Unlink()) {
		t.Error("Unlink() twice: want a StateError")
	}

	r := newLink()
	if err := r.Reject(now); err != nil || !r.Rejected() || r.Pending() {
		t.Errorf("Reject() = %v, rejected %v", err, r.Rejected())
	}
	if !isState(r.Unlink()) {
		t.Error("Unlink() of a rejected link: want a StateError")
	}
}

func TestLinkStatusService(t *testing.T) {
	f := NewMemberLinkFactory(&seqIDs{})
	yes, no := true, false
	pendingLink := func(at time.Time) *MemberLink { return f.Reconstruct("l", "a1", "m1", "i", nil, false, at, nil) }
	approvedLink := func(at time.Time) *MemberLink { return f.Reconstruct("l", "a1", "m1", "i", &yes, false, at, &at) }
	rejectedLink := func(at time.Time) *MemberLink { return f.Reconstruct("l", "a1", "m1", "i", &no, false, at, &at) }
	unlinked := func(at time.Time) *MemberLink { return f.Reconstruct("l", "a1", "m1", "i", &yes, true, at, &at) }
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
	for _, tt := range callerTests {
		t.Run(tt.name, func(t *testing.T) {
			if got := svc.CallerStatus(tt.links); got != tt.want {
				t.Errorf("CallerStatus() = %s, want %s", got, tt.want)
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
