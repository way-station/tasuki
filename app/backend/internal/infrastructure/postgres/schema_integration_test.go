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

func countRows(t *testing.T, client *Client, sql string, args ...any) int {
	t.Helper()
	var count int
	if err := client.Pool().QueryRow(context.Background(), sql, args...).Scan(&count); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
	return count
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

// saveSchemaRunningOrder saves a race with Taro and Jiro invited and in its entry,
// and a running order where Taro runs legs 1 and 3 and Jiro runs leg 2.
func saveSchemaRunningOrder(t *testing.T, client *Client) {
	t.Helper()
	saveSchemaRoster(t, client)
	mustExec(t, client, `INSERT INTO races (id, team_id, name, description) VALUES
		($1, $3, 'Hakone', '{}'), ($2, $3, 'Other race', '{}')`, schemaRaceID, schemaOtherRace, schemaTeamID)
	mustExec(t, client, "INSERT INTO race_invitations (race_id, roster_member_id) VALUES ($1, $2), ($1, $3)",
		schemaRaceID, schemaTaroID, schemaJiroID)
	mustExec(t, client, "INSERT INTO attendances (race_id, roster_member_id, status) VALUES ($1, $2, 'attending'), ($1, $3, 'attending')",
		schemaRaceID, schemaTaroID, schemaJiroID)
	mustExec(t, client, "INSERT INTO entries (id, race_id, name) VALUES ($1, $2, 'Team A')", schemaEntryID, schemaRaceID)
	mustExec(t, client, "INSERT INTO entry_members (entry_id, race_id, roster_member_id) VALUES ($1, $2, $3), ($1, $2, $4)",
		schemaEntryID, schemaRaceID, schemaTaroID, schemaJiroID)
	mustExec(t, client, "INSERT INTO running_orders (id, entry_id, status) VALUES ($1, $2, 'draft')", schemaOrderID, schemaEntryID)
	mustExec(t, client, `INSERT INTO running_order_legs (running_order_id, entry_id, leg_number, roster_member_id) VALUES
		($1, $2, 1, $3), ($1, $2, 2, $4), ($1, $2, 3, $3)`, schemaOrderID, schemaEntryID, schemaTaroID, schemaJiroID)
}

func TestCancellingAnInvitationRemovesAttendanceEntryAndLegs(t *testing.T) {
	client := newTestClient(t)
	saveSchemaRunningOrder(t, client)

	mustExec(t, client, "DELETE FROM race_invitations WHERE race_id = $1 AND roster_member_id = $2", schemaRaceID, schemaTaroID)

	if count := countRows(t, client, "SELECT count(*) FROM attendances WHERE roster_member_id = $1", schemaTaroID); count != 0 {
		t.Errorf("attendances of Taro = %d, want 0", count)
	}
	if count := countRows(t, client, "SELECT count(*) FROM entry_members WHERE roster_member_id = $1", schemaTaroID); count != 0 {
		t.Errorf("entry memberships of Taro = %d, want 0", count)
	}
	if count := countRows(t, client, "SELECT count(*) FROM running_order_legs WHERE roster_member_id IS NULL"); count != 2 {
		t.Errorf("empty legs = %d, want 2 (Taro's legs 1 and 3)", count)
	}
	if count := countRows(t, client, "SELECT count(*) FROM running_order_legs WHERE roster_member_id = $1", schemaJiroID); count != 1 {
		t.Errorf("legs of Jiro = %d, want 1", count)
	}
}

func TestRemovingAnEntryMemberEmptiesTheirLegs(t *testing.T) {
	client := newTestClient(t)
	saveSchemaRunningOrder(t, client)

	mustExec(t, client, "DELETE FROM entry_members WHERE entry_id = $1 AND roster_member_id = $2", schemaEntryID, schemaJiroID)

	if count := countRows(t, client, "SELECT count(*) FROM running_order_legs WHERE leg_number = 2 AND roster_member_id IS NULL"); count != 1 {
		t.Errorf("leg 2 is not empty")
	}
	if count := countRows(t, client, "SELECT count(*) FROM running_order_legs"); count != 3 {
		t.Errorf("legs = %d, want all 3 kept", count)
	}
}

func TestEntryAndRunningOrderKeys(t *testing.T) {
	client := newTestClient(t)
	saveSchemaRunningOrder(t, client)

	// A leg's runner must be a member of the same entry.
	mustExec(t, client, "DELETE FROM entry_members WHERE entry_id = $1 AND roster_member_id = $2", schemaEntryID, schemaJiroID)
	if code := execCode(t, client, "UPDATE running_order_legs SET roster_member_id = $1 WHERE leg_number = 2", schemaJiroID); code != foreignKeyFailed {
		t.Errorf("runner outside the entry: code %q, want %s", code, foreignKeyFailed)
	}
	// Only members invited to the entry's race can join it: Jiro is not invited
	// to the other race.
	if code := execCode(t, client, "INSERT INTO entry_members (entry_id, race_id, roster_member_id) VALUES ($1, $2, $3)",
		schemaEntryID, schemaOtherRace, schemaJiroID); code != foreignKeyFailed {
		t.Errorf("entry member of another race: code %q, want %s", code, foreignKeyFailed)
	}
	// One running order per entry.
	if code := execCode(t, client, "INSERT INTO running_orders (id, entry_id, status) VALUES (gen_random_uuid(), $1, 'draft')",
		schemaEntryID); code != uniqueViolation {
		t.Errorf("second running order: code %q, want %s", code, uniqueViolation)
	}
	// Deleting the running order deletes its legs.
	mustExec(t, client, "DELETE FROM running_orders WHERE id = $1", schemaOrderID)
	if count := countRows(t, client, "SELECT count(*) FROM running_order_legs"); count != 0 {
		t.Errorf("legs after deleting the running order = %d, want 0", count)
	}
}
