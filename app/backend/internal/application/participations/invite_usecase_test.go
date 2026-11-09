package participations

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestInviteUseCase(t *testing.T) {
	leftAt := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	memberFactory := roster.NewMemberFactory(fixedIDs{})
	members := &fakeMemberRepository{stored: map[shared.MemberID]*roster.Member{
		"m1": memberFactory.Reconstruct("m1", teamID, "Taro", nil),
		"m2": memberFactory.Reconstruct("m2", teamID, "Jiro", &leftAt),
		"m9": memberFactory.Reconstruct("m9", "t2", "Kuro", nil),
	}}
	invite := func(races *fakeRaceRepository, invitations *fakeInvitationRepository, memberID shared.MemberID) error {
		command := ParticipationCommandFactory{}.Invitation("r1", memberID)
		return NewInviteUseCase(races, members, invitations, participation.InvitationFactory{}, teamID).Execute(context.Background(), command)
	}

	t.Run("saves the invitation", func(t *testing.T) {
		invitations := invitedTo("r1")
		if err := invite(&fakeRaceRepository{stored: teamRace(t, teamID)}, invitations, "m1"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if invitations.saved == nil || invitations.saved.RaceID() != "r1" || invitations.saved.MemberID() != "m1" {
			t.Errorf("saved %+v, want m1 to r1", invitations.saved)
		}
	})

	t.Run("does nothing for an invited member", func(t *testing.T) {
		invitations := invitedTo("r1", "m1")
		if err := invite(&fakeRaceRepository{stored: teamRace(t, teamID)}, invitations, "m1"); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if invitations.saved != nil {
			t.Error("the invitation was saved again")
		}
	})

	t.Run("rejects a member who has left", func(t *testing.T) {
		invitations := invitedTo("r1")
		if err := invite(&fakeRaceRepository{stored: teamRace(t, teamID)}, invitations, "m2"); !isValidation(err) {
			t.Errorf("Execute() error = %v, want a validation error", err)
		}
		if invitations.saved != nil {
			t.Error("a member who has left was invited")
		}
	})

	failures := []struct {
		name     string
		races    *fakeRaceRepository
		memberID shared.MemberID
		findErr  error
		want     error
	}{
		{name: "the race does not exist", races: &fakeRaceRepository{}, memberID: "m1", want: shared.ErrNotFound},
		{name: "the race belongs to another team", races: &fakeRaceRepository{stored: teamRace(t, "t2")}, memberID: "m1", want: shared.ErrNotFound},
		{name: "the member does not exist", races: &fakeRaceRepository{stored: teamRace(t, teamID)}, memberID: "m8", want: shared.ErrNotFound},
		{name: "the member belongs to another team", races: &fakeRaceRepository{stored: teamRace(t, teamID)}, memberID: "m9", want: shared.ErrNotFound},
		{name: "the invitation cannot be read", races: &fakeRaceRepository{stored: teamRace(t, teamID)}, memberID: "m1", findErr: errDatabase, want: errDatabase},
	}
	for _, testCase := range failures {
		t.Run(testCase.name, func(t *testing.T) {
			invitations := invitedTo("r1")
			invitations.findErr = testCase.findErr
			if err := invite(testCase.races, invitations, testCase.memberID); !errors.Is(err, testCase.want) {
				t.Errorf("Execute() error = %v, want %v", err, testCase.want)
			}
			if invitations.saved != nil {
				t.Error("the invitation was saved")
			}
		})
	}
}
