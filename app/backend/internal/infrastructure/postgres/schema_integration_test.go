//go:build integration

package postgres

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

// Rules the schema enforces by itself (docs/design/database.md): partial unique
// indexes, and rows that foreign keys delete or empty.

const (
	schemaTeamID     = "00000000-0000-7000-8000-000000000101"
	schemaTaroID     = "00000000-0000-7000-8000-000000000111"
	schemaJiroID     = "00000000-0000-7000-8000-000000000112"
	schemaUserID     = "00000000-0000-7000-8000-000000000121"
	schemaOtherUser  = "00000000-0000-7000-8000-000000000122"
	schemaRaceID     = "00000000-0000-7000-8000-000000000131"
	schemaOtherRace  = "00000000-0000-7000-8000-000000000132"
	schemaEntryID    = "00000000-0000-7000-8000-000000000141"
	schemaOrderID    = "00000000-0000-7000-8000-000000000151"
	uniqueViolation  = "23505"
	foreignKeyFailed = "23503"
)

// execCode runs sql and returns the PostgreSQL error code, or "" when it succeeds.
func execCode(t *testing.T, client *Client, sql string, args ...any) string {
	t.Helper()
	_, err := client.Pool().Exec(context.Background(), sql, args...)
	if err == nil {
		return ""
	}
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		t.Fatalf("%s: %v", sql, err)
	}
	return postgresError.Code
}

// saveSchemaRoster saves a team with two roster members, Taro and Jiro.
func saveSchemaRoster(t *testing.T, client *Client) {
	t.Helper()
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Team')", schemaTeamID)
	mustExec(t, client, "INSERT INTO roster_members (id, team_id, name) VALUES ($1, $3, 'Taro'), ($2, $3, 'Jiro')",
		schemaTaroID, schemaJiroID, schemaTeamID)
}

func TestInviteLinksOneOpenPerMember(t *testing.T) {
	client := newTestClient(t)
	saveSchemaRoster(t, client)
	insert := "INSERT INTO invite_links (id, roster_member_id, token_hash, expires_at, used_at, revoked_at) VALUES (gen_random_uuid(), $1, $2, now(), $3, $4)"

	if code := execCode(t, client, insert, schemaTaroID, []byte("1"), nil, nil); code != "" {
		t.Fatalf("first open link: %s", code)
	}
	if code := execCode(t, client, insert, schemaTaroID, []byte("2"), nil, nil); code != uniqueViolation {
		t.Errorf("second open link: code %q, want %s", code, uniqueViolation)
	}
	if code := execCode(t, client, insert, schemaTaroID, []byte("3"), nil, testNow); code != "" {
		t.Errorf("replaced link next to an open one: %s", code)
	}
	if code := execCode(t, client, insert, schemaJiroID, []byte("4"), nil, nil); code != "" {
		t.Errorf("open link of another member: %s", code)
	}
}

func TestMemberLinksOneApprovedAndOnePending(t *testing.T) {
	client := newTestClient(t)
	saveSchemaRoster(t, client)
	mustExec(t, client, "INSERT INTO users (id, line_user_id, display_name) VALUES ($1, 'U-1', 'Taro'), ($2, 'U-2', 'Other')",
		schemaUserID, schemaOtherUser)
	linkNumber := 0
	// Each member link needs its own invite link.
	insertLink := func(userID, rosterMemberID string, approved, rejected bool) string {
		t.Helper()
		linkNumber++
		inviteLinkID := ""
		if err := client.Pool().QueryRow(context.Background(),
			"INSERT INTO invite_links (id, roster_member_id, token_hash, expires_at, used_at) VALUES (gen_random_uuid(), $1, $2, now(), now()) RETURNING id",
			rosterMemberID, []byte(fmt.Sprintf("link-%d", linkNumber))).Scan(&inviteLinkID); err != nil {
			t.Fatalf("invite link: %v", err)
		}
		var approvedAt, rejectedAt any
		if approved {
			approvedAt = testNow
		}
		if rejected {
			rejectedAt = testNow
		}
		return execCode(t, client,
			"INSERT INTO member_links (id, user_id, roster_member_id, invite_link_id, approved_at, rejected_at) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)",
			userID, rosterMemberID, inviteLinkID, approvedAt, rejectedAt)
	}

	if code := insertLink(schemaUserID, schemaTaroID, false, false); code != "" {
		t.Fatalf("pending link: %s", code)
	}
	if code := insertLink(schemaOtherUser, schemaTaroID, false, false); code != uniqueViolation {
		t.Errorf("second pending link for the member: code %q, want %s", code, uniqueViolation)
	}
	if code := insertLink(schemaUserID, schemaJiroID, false, false); code != uniqueViolation {
		t.Errorf("second pending link for the user: code %q, want %s", code, uniqueViolation)
	}
	if code := insertLink(schemaOtherUser, schemaJiroID, false, true); code != "" {
		t.Errorf("rejected link: %s", code)
	}

	mustExec(t, client, "UPDATE member_links SET approved_at = now() WHERE user_id = $1", schemaUserID)
	if code := insertLink(schemaOtherUser, schemaTaroID, true, false); code != uniqueViolation {
		t.Errorf("second approved user for the member: code %q, want %s", code, uniqueViolation)
	}
	if code := insertLink(schemaUserID, schemaJiroID, true, false); code != uniqueViolation {
		t.Errorf("second approved member for the user: code %q, want %s", code, uniqueViolation)
	}

	mustExec(t, client, "UPDATE member_links SET unlinked_at = now() WHERE user_id = $1", schemaUserID)
	if code := insertLink(schemaOtherUser, schemaTaroID, true, false); code != "" {
		t.Errorf("approved link after unlinking: %s", code)
	}
}
