package roster

import (
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fixedIDs struct{ id string }

func (ids fixedIDs) NewID() string { return ids.id }

func newMember(t *testing.T) *Member {
	t.Helper()
	member, err := NewMemberFactory(fixedIDs{id: "m1"}).New("team1", "  Sato  ")
	if err != nil {
		t.Fatal(err)
	}
	return member
}

func TestMemberFactoryNew(t *testing.T) {
	member := newMember(t)
	if member.ID() != "m1" || member.TeamID() != "team1" || member.Name() != "Sato" || member.Left() {
		t.Errorf("New() = %+v", member)
	}

	_, err := NewMemberFactory(fixedIDs{id: "m2"}).New("team1", " ")
	var validationErr *shared.ValidationError
	if !errors.As(err, &validationErr) {
		t.Errorf("New(blank name) error = %v, want a ValidationError", err)
	}
}

func TestLeaveAndRejoin(t *testing.T) {
	member := newMember(t)
	first := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	member.Leave(first)
	member.Leave(first.Add(time.Hour)) // does nothing
	if leftAt, ok := member.LeftAt(); !ok || !leftAt.Equal(first) {
		t.Errorf("LeftAt() = %v, %v, want %v", leftAt, ok, first)
	}
	member.Rejoin()
	if member.Left() {
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
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := testCase.member.CanIssueInviteLink(testCase.link)
			var stateErr *shared.StateError
			if testCase.wantErr != errors.As(err, &stateErr) {
				t.Errorf("CanIssueInviteLink() = %v, want error: %v", err, testCase.wantErr)
			}
		})
	}
}

func TestInvitationsToCancel(t *testing.T) {
	date := func(text string) shared.Date {
		parsed, err := shared.ParseDate(text)
		if err != nil {
			t.Fatal(err)
		}
		return parsed
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
