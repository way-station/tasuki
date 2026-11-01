//go:build integration

package postgres

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

// takeConcurrently calls take from two goroutines at once and returns how many succeeded.
func takeConcurrently(take func() error) int {
	var wait sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wait.Go(func() { results <- take() })
	}
	wait.Wait()
	close(results)
	succeeded := 0
	for err := range results {
		if err == nil {
			succeeded++
		}
	}
	return succeeded
}

func saveTestInviteLink(t *testing.T, client *Client) *identity.InviteLink {
	t.Helper()
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ('00000000-0000-7000-8000-000000000001', 'Team')")
	mustExec(t, client, `INSERT INTO roster_members (id, team_id, name) VALUES
		('00000000-0000-7000-8000-000000000011', '00000000-0000-7000-8000-000000000001', 'Taro Yamada')`)
	mustExec(t, client, `INSERT INTO invite_links (id, roster_member_id, token_hash, expires_at) VALUES
		('00000000-0000-7000-8000-000000000021', '00000000-0000-7000-8000-000000000011', 'a', $1)`,
		testNow.Add(identity.InviteLinkLifetime))
	return identity.NewInviteLinkFactory(random.IDGenerator{}, random.TokenGenerator{}).Reconstruct(
		"00000000-0000-7000-8000-000000000021", "00000000-0000-7000-8000-000000000011",
		[]byte("a"), testNow.Add(identity.InviteLinkLifetime), nil, nil)
}

func TestLoginAttemptRepository(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	factory := identity.NewLoginAttemptFactory(random.IDGenerator{}, random.TokenGenerator{})
	repository := NewLoginAttemptRepository(client, factory)
	inviteLink := saveTestInviteLink(t, client)

	t.Run("takes an attempt once, with its invite link", func(t *testing.T) {
		attempt, state, err := factory.New(inviteLink, testNow)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, attempt); err != nil {
			t.Fatalf("Save: %v", err)
		}

		taken, err := repository.Take(ctx, identity.HashToken(state))
		if err != nil {
			t.Fatalf("Take: %v", err)
		}
		inviteLinkID, ok := taken.InviteLinkID()
		if taken.ID() != attempt.ID() || !bytes.Equal(taken.StateHash(), attempt.StateHash()) ||
			taken.Nonce() != attempt.Nonce() || taken.CodeVerifier() != attempt.CodeVerifier() ||
			!ok || inviteLinkID != inviteLink.ID() || !taken.ExpiresAt().Equal(attempt.ExpiresAt()) {
			t.Errorf("Take = %+v, want %+v", taken, attempt)
		}

		if _, err := repository.Take(ctx, identity.HashToken(state)); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Take again error = %v, want ErrNotFound", err)
		}
	})

	t.Run("takes an attempt without an invite link", func(t *testing.T) {
		attempt, state, err := factory.New(nil, testNow)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, attempt); err != nil {
			t.Fatalf("Save: %v", err)
		}

		taken, err := repository.Take(ctx, identity.HashToken(state))
		if _, ok := taken.InviteLinkID(); err != nil || ok {
			t.Errorf("Take = %+v, %v, want no invite link", taken, err)
		}
	})

	t.Run("lets only one of two callers take an attempt", func(t *testing.T) {
		attempt, state, err := factory.New(nil, testNow)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, attempt); err != nil {
			t.Fatalf("Save: %v", err)
		}

		succeeded := takeConcurrently(func() error {
			_, err := repository.Take(ctx, identity.HashToken(state))
			return err
		})

		if succeeded != 1 {
			t.Errorf("%d callers took the attempt, want 1", succeeded)
		}
	})
}

func TestLoginCodeRepository(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	user := saveTestUser(t, client)
	factory := identity.NewLoginCodeFactory(random.IDGenerator{}, random.TokenGenerator{})
	repository := NewLoginCodeRepository(client, factory)

	t.Run("takes a code once", func(t *testing.T) {
		code, plain, err := factory.New(user.ID(), testNow)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, code); err != nil {
			t.Fatalf("Save: %v", err)
		}

		taken, err := repository.Take(ctx, identity.HashToken(plain))
		if err != nil || taken.ID() != code.ID() || taken.UserID() != user.ID() ||
			!bytes.Equal(taken.CodeHash(), code.CodeHash()) || !taken.ExpiresAt().Equal(code.ExpiresAt()) {
			t.Fatalf("Take = %+v, %v, want %+v", taken, err, code)
		}

		if _, err := repository.Take(ctx, identity.HashToken(plain)); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Take again error = %v, want ErrNotFound", err)
		}
	})

	t.Run("lets only one of two callers take a code", func(t *testing.T) {
		code, plain, err := factory.New(user.ID(), testNow)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, code); err != nil {
			t.Fatalf("Save: %v", err)
		}

		succeeded := takeConcurrently(func() error {
			_, err := repository.Take(ctx, identity.HashToken(plain))
			return err
		})

		if succeeded != 1 {
			t.Errorf("%d callers took the code, want 1", succeeded)
		}
	})
}
