package rosters

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fixedToken string

func (token fixedToken) NewToken() (string, error) { return string(token), nil }

// fakeInviteLinks keeps invite links by ID.
type fakeInviteLinks struct {
	stored map[shared.InviteLinkID]*identity.InviteLink
}

func (fake *fakeInviteLinks) FindByTokenHash(context.Context, []byte) (*identity.InviteLink, error) {
	return nil, errors.New("invite links are not found by token in this package")
}

func (fake *fakeInviteLinks) FindOpenByMember(_ context.Context, memberID shared.MemberID) (*identity.InviteLink, error) {
	for _, invite := range fake.stored {
		_, used := invite.UsedAt()
		_, revoked := invite.RevokedAt()
		if invite.MemberID() == memberID && !used && !revoked {
			return invite, nil
		}
	}
	return nil, fmt.Errorf("open invite link of %s: %w", memberID, shared.ErrNotFound)
}

func (fake *fakeInviteLinks) Save(ctx context.Context, invite *identity.InviteLink) error {
	if !inTransaction(ctx) {
		return errors.New("the invite link was saved outside a unit of work")
	}
	fake.stored[invite.ID()] = invite
	return nil
}

func TestCreateInviteLinkUseCase(t *testing.T) {
	newUseCase := func(member *roster.Member, links *fakeLinkRepository, invites *fakeInviteLinks) *CreateInviteLinkUseCase {
		factory := identity.NewInviteLinkFactory(&sequentialIDs{}, fixedToken("token"))
		return NewCreateInviteLinkUseCase(fakeUnitOfWork{}, membersOf(member), links, invites, factory, identity.LinkStatusService{}, RosterDTOFactory{}, clock, teamID)
	}

	t.Run("replaces the previous link", func(t *testing.T) {
		previous := identity.NewInviteLinkFactory(nil, nil).Reconstruct("old", "m1", nil, testNow.Add(time.Hour), nil, nil)
		invites := &fakeInviteLinks{stored: map[shared.InviteLinkID]*identity.InviteLink{"old": previous}}
		created, err := newUseCase(teamMember("m1", teamID), linksOf(link("k1", "u1", "m1", false, true)), invites).Execute(context.Background(), "m1")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if created.Token() != "token" || !created.ExpiresAt().Equal(testNow.Add(identity.InviteLinkLifetime)) {
			t.Errorf("Execute() = %+v, want the token for a week", created)
		}
		if _, revoked := invites.stored["old"].RevokedAt(); !revoked || len(invites.stored) != 2 {
			t.Errorf("invite links = %v, want the old one revoked and a new one", invites.stored)
		}
	})

	left := teamMember("m1", teamID)
	left.Leave(testNow)
	tests := []struct {
		name   string
		member *roster.Member
		links  *fakeLinkRepository
	}{
		{name: "a member who has left", member: left, links: linksOf()},
		{name: "a member with an approved link", member: teamMember("m1", teamID), links: linksOf(link("k1", "u1", "m1", true, false))},
		{name: "a member with a pending link", member: teamMember("m1", teamID), links: linksOf(link("k1", "u1", "m1", false, false))},
	}
	for _, testCase := range tests {
		t.Run("rejects "+testCase.name, func(t *testing.T) {
			invites := &fakeInviteLinks{stored: map[shared.InviteLinkID]*identity.InviteLink{}}
			_, err := newUseCase(testCase.member, testCase.links, invites).Execute(context.Background(), "m1")
			var stateErr *shared.StateError
			if !errors.As(err, &stateErr) || len(invites.stored) != 0 {
				t.Errorf("Execute() error = %v, want a state error and no link", err)
			}
		})
	}
}
