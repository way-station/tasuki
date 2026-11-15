//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

func TestMemberLinkRepository(t *testing.T) {
	client := newTestClient(t)
	saveLinkFixtures(t, client)
	ctx := context.Background()
	invites := identity.NewInviteLinkFactory(random.IDGenerator{}, fixedLinkToken("token"))
	invite, _, err := invites.New(linkTaroID, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewInviteLinkRepository(client, invites).Save(ctx, invite); err != nil {
		t.Fatal(err)
	}
	factory := identity.NewMemberLinkFactory(random.IDGenerator{})
	repository := NewMemberLinkRepository(client, factory)

	link, err := factory.New(linkUserID, invite, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, link); err != nil {
		t.Fatalf("Save: %v", err)
	}
	found, err := repository.FindByID(ctx, link.ID())
	if err != nil || !found.Pending() || found.UserID() != linkUserID || found.MemberID() != linkTaroID || found.InviteLinkID() != invite.ID() || !found.CreatedAt().Equal(testNow) {
		t.Fatalf("FindByID = %+v, %v, want the pending link", found, err)
	}

	t.Run("stores approving and unlinking", func(t *testing.T) {
		if err := found.Approve(testNow.Add(time.Hour)); err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, found); err != nil {
			t.Fatalf("Save approved: %v", err)
		}
		if err := found.Unlink(testNow.Add(2 * time.Hour)); err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, found); err != nil {
			t.Fatalf("Save unlinked: %v", err)
		}
		stored, err := repository.FindByID(ctx, link.ID())
		if _, approved := stored.ApprovedAt(); err != nil || !approved || !stored.Unlinked() {
			t.Errorf("FindByID = %+v, %v, want an approved link that was unlinked", stored, err)
		}
	})

	t.Run("lists the links of the user and of the member", func(t *testing.T) {
		byUser, err := repository.ListByUser(ctx, linkUserID)
		if err != nil || len(byUser) != 1 || byUser[0].ID() != link.ID() {
			t.Errorf("ListByUser = %v, %v, want the link", byUser, err)
		}
		byMember, err := repository.ListByMember(ctx, linkTaroID)
		if err != nil || len(byMember) != 1 || byMember[0].ID() != link.ID() {
			t.Errorf("ListByMember = %v, %v, want the link", byMember, err)
		}
	})

	if _, err := repository.FindByID(ctx, shared.MemberLinkID(random.IDGenerator{}.NewID())); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindByID(unknown) error = %v, want ErrNotFound", err)
	}
}
