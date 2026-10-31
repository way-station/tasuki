//go:build integration

package postgres

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

// newTestClient connects to the migrated test database and empties the tables the tests use.
func newTestClient(t *testing.T) *Client {
	t.Helper()
	ctx := context.Background()
	client, err := Connect(ctx, testDatabaseURL(t))
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	t.Cleanup(client.Close)
	if _, err := client.Pool().Exec(ctx, "TRUNCATE sessions, member_links, invite_links, roster_members, users, teams CASCADE"); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	return client
}

func mustExec(t *testing.T, client *Client, sql string, args ...any) {
	t.Helper()
	if _, err := client.Pool().Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

var testNow = time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC)

func TestUserRepository(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	factory := identity.NewUserFactory(random.IDGenerator{})
	repository := NewUserRepository(client, factory)

	user, err := factory.New("U-1", "Taro")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, user); err != nil {
		t.Fatalf("Save: %v", err)
	}

	found, err := repository.FindByLineUserID(ctx, "U-1")
	if err != nil || found.ID() != user.ID() || found.DisplayName() != "Taro" {
		t.Fatalf("FindByLineUserID = %v, %v", found, err)
	}

	found.UpdateDisplayName("Taro Y")
	if err := repository.Save(ctx, found); err != nil {
		t.Fatalf("Save again: %v", err)
	}
	byID, err := repository.FindByID(ctx, user.ID())
	if err != nil || byID.DisplayName() != "Taro Y" {
		t.Errorf("FindByID after renaming = %v, %v", byID, err)
	}

	if _, err := repository.FindByLineUserID(ctx, "U-unknown"); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindByLineUserID(unknown) error = %v, want ErrNotFound", err)
	}
}

func saveTestUser(t *testing.T, client *Client) *identity.User {
	t.Helper()
	user, err := identity.NewUserFactory(random.IDGenerator{}).New("U-1", "Taro")
	if err != nil {
		t.Fatal(err)
	}
	if err := NewUserRepository(client, identity.NewUserFactory(random.IDGenerator{})).Save(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	return user
}

func TestSessionRepositoryAndQueryService(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	user := saveTestUser(t, client)
	factory := identity.NewSessionFactory(random.IDGenerator{}, random.TokenGenerator{})
	repository := NewSessionRepository(client, factory)
	queryService := NewSessionQueryService(client)

	session, token, err := factory.New(user.ID(), testNow)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, session); err != nil {
		t.Fatalf("Save: %v", err)
	}
	hash := identity.HashToken(token)

	found, err := repository.FindByTokenHash(ctx, hash)
	if err != nil || found.ID() != session.ID() || !bytes.Equal(found.TokenHash(), hash) || !found.ExpiresAt().Equal(session.ExpiresAt()) {
		t.Fatalf("FindByTokenHash = %v, %v", found, err)
	}

	if userID, err := queryService.FindActiveUserID(ctx, hash, testNow); err != nil || userID != user.ID() {
		t.Errorf("FindActiveUserID(active) = %q, %v, want %q", userID, err, user.ID())
	}
	if _, err := queryService.FindActiveUserID(ctx, hash, session.ExpiresAt()); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindActiveUserID(expired) error = %v, want ErrNotFound", err)
	}
	if _, err := queryService.FindActiveUserID(ctx, identity.HashToken("unknown"), testNow); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindActiveUserID(unknown) error = %v, want ErrNotFound", err)
	}

	found.Revoke(testNow.Add(time.Hour))
	if err := repository.Save(ctx, found); err != nil {
		t.Fatalf("Save revoked: %v", err)
	}
	if _, err := queryService.FindActiveUserID(ctx, hash, testNow.Add(2*time.Hour)); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindActiveUserID(revoked) error = %v, want ErrNotFound", err)
	}
	revoked, err := repository.FindByTokenHash(ctx, hash)
	if at, ok := revoked.RevokedAt(); err != nil || !ok || !at.Equal(testNow.Add(time.Hour)) {
		t.Errorf("RevokedAt() = %v, %v, %v", at, ok, err)
	}
}

func TestMeQueryService(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()
	user := saveTestUser(t, client)
	queryService := NewMeQueryService(client)

	t.Run("a user without links", func(t *testing.T) {
		record, err := queryService.FindMe(ctx, user.ID())
		if err != nil || record.LineUserID != "U-1" || record.DisplayName != "Taro" || len(record.Links) != 0 {
			t.Fatalf("FindMe = %+v, %v", record, err)
		}
	})

	t.Run("links, oldest first, with the roster member's name", func(t *testing.T) {
		mustExec(t, client, "INSERT INTO teams (id, name) VALUES ('00000000-0000-7000-8000-000000000001', 'Team')")
		mustExec(t, client, `INSERT INTO roster_members (id, team_id, name) VALUES
			('00000000-0000-7000-8000-000000000011', '00000000-0000-7000-8000-000000000001', 'Taro Yamada'),
			('00000000-0000-7000-8000-000000000012', '00000000-0000-7000-8000-000000000001', 'Jiro')`)
		mustExec(t, client, `INSERT INTO invite_links (id, roster_member_id, token_hash, expires_at, used_at) VALUES
			('00000000-0000-7000-8000-000000000021', '00000000-0000-7000-8000-000000000012', 'a', $1, $1),
			('00000000-0000-7000-8000-000000000022', '00000000-0000-7000-8000-000000000011', 'b', $1, $1)`, testNow)
		mustExec(t, client, `INSERT INTO member_links (id, user_id, roster_member_id, invite_link_id, created_at, rejected_at) VALUES
			('00000000-0000-7000-8000-000000000031', $1, '00000000-0000-7000-8000-000000000012', '00000000-0000-7000-8000-000000000021', $2, $2)`,
			string(user.ID()), testNow)
		mustExec(t, client, `INSERT INTO member_links (id, user_id, roster_member_id, invite_link_id, created_at, approved_at) VALUES
			('00000000-0000-7000-8000-000000000032', $1, '00000000-0000-7000-8000-000000000011', '00000000-0000-7000-8000-000000000022', $2, $2)`,
			string(user.ID()), testNow.Add(time.Hour))

		record, err := queryService.FindMe(ctx, user.ID())
		if err != nil || len(record.Links) != 2 {
			t.Fatalf("FindMe = %+v, %v", record, err)
		}
		rejected, approved := record.Links[0], record.Links[1]
		if rejected.RosterMemberName != "Jiro" || rejected.RejectedAt == nil || rejected.ApprovedAt != nil {
			t.Errorf("first link = %+v, want the rejected link to Jiro", rejected)
		}
		if approved.RosterMemberName != "Taro Yamada" || approved.ApprovedAt == nil || approved.RosterMemberID != "00000000-0000-7000-8000-000000000011" {
			t.Errorf("second link = %+v, want the approved link to Taro Yamada", approved)
		}
	})

	t.Run("a user that does not exist", func(t *testing.T) {
		if _, err := queryService.FindMe(ctx, shared.UserID(random.IDGenerator{}.NewID())); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("FindMe error = %v, want ErrNotFound", err)
		}
	})
}
