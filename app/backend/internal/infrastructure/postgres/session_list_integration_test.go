//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

func TestSessionRepositoryListActiveByUser(t *testing.T) {
	client := newTestClient(t)
	saveLinkFixtures(t, client)
	ctx := context.Background()
	factory := identity.NewSessionFactory(random.IDGenerator{}, random.TokenGenerator{})
	repository := NewSessionRepository(client, factory)
	var sessions []*identity.Session
	for range 3 {
		session, _, err := factory.New(linkUserID, testNow)
		if err != nil {
			t.Fatal(err)
		}
		sessions = append(sessions, session)
	}
	sessions[1].Revoke(testNow)
	for _, session := range sessions {
		if err := repository.Save(ctx, session); err != nil {
			t.Fatal(err)
		}
	}

	active, err := repository.ListActiveByUser(ctx, linkUserID, testNow.Add(time.Minute))
	if err != nil || len(active) != 2 {
		t.Fatalf("ListActiveByUser = %v, %v, want the two sessions that were not revoked", active, err)
	}
	expired, err := repository.ListActiveByUser(ctx, linkUserID, sessions[0].ExpiresAt().Add(time.Minute))
	if err != nil || len(expired) != 0 {
		t.Errorf("ListActiveByUser after the expiry = %v, %v, want none", expired, err)
	}

	t.Run("saves in the transaction of a unit of work", func(t *testing.T) {
		rolledBack := errors.New("roll back")
		err := NewUnitOfWork(client).Run(ctx, func(ctx context.Context) error {
			active[0].Revoke(testNow.Add(time.Minute))
			if err := repository.Save(ctx, active[0]); err != nil {
				return err
			}
			return rolledBack
		})
		if !errors.Is(err, rolledBack) {
			t.Fatalf("Run error = %v", err)
		}
		if still, err := repository.ListActiveByUser(ctx, linkUserID, testNow.Add(time.Minute)); err != nil || len(still) != 2 {
			t.Errorf("ListActiveByUser after a rollback = %v, %v, want both sessions still active", still, err)
		}
	})
}
