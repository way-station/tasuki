//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

func TestInviteLinkQueryService(t *testing.T) {
	client := newTestClient(t)
	saveLinkFixtures(t, client)
	ctx := context.Background()
	factory := identity.NewInviteLinkFactory(random.IDGenerator{}, fixedLinkToken("query"))
	link, token, err := factory.New(linkTaroID, testNow)
	if err != nil {
		t.Fatal(err)
	}
	link.Revoke(testNow.Add(time.Minute))
	if err := NewInviteLinkRepository(client, factory).Save(ctx, link); err != nil {
		t.Fatalf("Save: %v", err)
	}
	service := NewInviteLinkQueryService(client)

	check := func(t *testing.T, record auth.InviteLinkRecord) {
		t.Helper()
		if record.ID != link.ID() || record.MemberID != linkTaroID || string(record.TokenHash) != string(link.TokenHash()) || !record.ExpiresAt.Equal(link.ExpiresAt()) {
			t.Errorf("record = %+v, want the saved link", record)
		}
		if record.UsedAt != nil || record.RevokedAt == nil || !record.RevokedAt.Equal(testNow.Add(time.Minute)) {
			t.Errorf("UsedAt, RevokedAt = %v, %v, want nil and the revocation", record.UsedAt, record.RevokedAt)
		}
		if record.TeamName != "Team" {
			t.Errorf("TeamName = %q, want Team", record.TeamName)
		}
	}

	t.Run("by the token hash", func(t *testing.T) {
		record, err := service.FindByTokenHash(ctx, identity.HashToken(token))
		if err != nil {
			t.Fatal(err)
		}
		check(t, record)
	})

	t.Run("by ID in a unit of work", func(t *testing.T) {
		err := NewUnitOfWork(client).Run(ctx, func(ctx context.Context) error {
			record, err := service.FindByID(ctx, link.ID())
			if err != nil {
				return err
			}
			check(t, record)
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
	})

	t.Run("unknown links", func(t *testing.T) {
		if _, err := service.FindByTokenHash(ctx, identity.HashToken("unknown")); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("FindByTokenHash(unknown) error = %v, want ErrNotFound", err)
		}
		if _, err := service.FindByID(ctx, "01900000-0000-7000-8000-000000000cff"); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("FindByID(unknown) error = %v, want ErrNotFound", err)
		}
	})
}
