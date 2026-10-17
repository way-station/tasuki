package roster

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fixedIDs struct{ id string }

func (f fixedIDs) NewID() string { return f.id }

func newMember(t *testing.T) *Member {
	t.Helper()
	m, err := NewMemberFactory(fixedIDs{id: "m1"}).New("team1", "  Sato  ")
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestMemberFactoryNew(t *testing.T) {
	m := newMember(t)
	if m.ID() != "m1" || m.TeamID() != "team1" || m.Name() != "Sato" || m.Left() {
		t.Errorf("New() = %+v", m)
	}

	_, err := NewMemberFactory(fixedIDs{id: "m2"}).New("team1", " ")
	var ve *shared.ValidationError
	if !errors.As(err, &ve) {
		t.Errorf("New(blank name) error = %v, want a ValidationError", err)
	}
}

func TestLeaveAndRejoin(t *testing.T) {
	m := newMember(t)
	first := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	m.Leave(first)
	m.Leave(first.Add(time.Hour)) // does nothing
	if at, ok := m.LeftAt(); !ok || !at.Equal(first) {
		t.Errorf("LeftAt() = %v, %v, want %v", at, ok, first)
	}
	m.Rejoin()
	if m.Left() {
		t.Error("Left() after Rejoin() = true")
	}
}

func TestCanIssueInviteLink(t *testing.T) {
	left := newMember(t)
	left.Leave(time.Now())

	tests := []struct {
		name    string
		member  *Member
		link    LinkStatus
		wantErr bool
	}{
		{name: "no link", member: newMember(t), link: LinkNone},
		{name: "pending", member: newMember(t), link: LinkPending, wantErr: true},
		{name: "approved", member: newMember(t), link: LinkApproved, wantErr: true},
		{name: "left", member: left, link: LinkNone, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.member.CanIssueInviteLink(tt.link)
			var se *shared.StateError
			if tt.wantErr != errors.As(err, &se) {
				t.Errorf("CanIssueInviteLink() = %v, want error: %v", err, tt.wantErr)
			}
		})
	}
}

func TestInvitationsToCancel(t *testing.T) {
	date := func(s string) shared.Date {
		d, err := shared.ParseDate(s)
		if err != nil {
			t.Fatal(err)
		}
		return d
	}
	got := LeaveService{}.InvitationsToCancel(date("2026-10-17"), []InvitedRace{
		{RaceID: "past", RaceDate: date("2026-10-16")},
		{RaceID: "today", RaceDate: date("2026-10-17")},
		{RaceID: "future", RaceDate: date("2026-11-23")},
	})
	if want := []shared.RaceID{"today", "future"}; !slices.Equal(got, want) {
		t.Errorf("InvitationsToCancel() = %v, want %v", got, want)
	}
}
