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

const (
	linkTeamID = "01900000-0000-7000-8000-000000000c01"
	linkTaroID = "01900000-0000-7000-8000-000000000c21"
	linkUserID = "01900000-0000-7000-8000-000000000c31"
)

// saveLinkFixtures saves a team with Taro and a user.
func saveLinkFixtures(t *testing.T, client *Client) {
	t.Helper()
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Team')", linkTeamID)
	mustExec(t, client, "INSERT INTO roster_members (id, team_id, name) VALUES ($1, $2, 'Taro')", linkTaroID, linkTeamID)
	mustExec(t, client, "INSERT INTO users (id, line_user_id, display_name) VALUES ($1, 'U-taro', 'taro')", linkUserID)
}

type fixedLinkToken string

func (token fixedLinkToken) NewToken() (string, error) { return string(token), nil }

func TestInviteLinkRepository(t *testing.T) {
	client := newTestClient(t)
	saveLinkFixtures(t, client)
	ctx := context.Background()
	factory := identity.NewInviteLinkFactory(random.IDGenerator{}, fixedLinkToken("first"))
	repository := NewInviteLinkRepository(client, factory)

	first, token, err := factory.New(linkTaroID, testNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, first); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repository.FindByTokenHash(ctx, identity.HashToken(token))
	if err != nil || found.ID() != first.ID() || found.MemberID() != linkTaroID || !found.ExpiresAt().Equal(first.ExpiresAt()) {
		t.Fatalf("FindByTokenHash = %+v, %v, want the first link", found, err)
	}
	if open, err := repository.FindOpenByMember(ctx, linkTaroID); err != nil || open.ID() != first.ID() {
		t.Errorf("FindOpenByMember = %+v, %v, want the first link", open, err)
	}

	t.Run("stores that a link was replaced", func(t *testing.T) {
		found.Revoke(testNow.Add(time.Minute))
		if err := repository.Save(ctx, found); err != nil {
			t.Fatalf("Save revoked: %v", err)
		}
		if _, err := repository.FindOpenByMember(ctx, linkTaroID); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("FindOpenByMember after revoking error = %v, want ErrNotFound", err)
		}
		revoked, err := repository.FindByTokenHash(ctx, identity.HashToken(token))
		if at, ok := revoked.RevokedAt(); err != nil || !ok || !at.Equal(testNow.Add(time.Minute)) {
			t.Errorf("RevokedAt = %v, %v, %v, want the time it was replaced", at, ok, err)
		}
	})

	if _, err := repository.FindByTokenHash(ctx, identity.HashToken("unknown")); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindByTokenHash(unknown) error = %v, want ErrNotFound", err)
	}
}
