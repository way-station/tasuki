//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

// saveTestMember stores a roster member of the team, who has left at leftAt unless it is
// nil, and returns their ID.
func saveTestMember(t *testing.T, client *Client, teamID shared.TeamID, name string, leftAt *time.Time) shared.MemberID {
	t.Helper()
	id := shared.MemberID(random.IDGenerator{}.NewID())
	mustExec(t, client, "INSERT INTO roster_members (id, team_id, name, left_at) VALUES ($1, $2, $3, $4)", string(id), string(teamID), name, leftAt)
	return id
}

func TestInvitationRepository(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	ctx := context.Background()
	repository := NewInvitationRepository(client, participation.InvitationFactory{})
	raceID, otherRaceID := saveTestRace(t, client, testTeamID, "Ekiden"), saveTestRace(t, client, testTeamID, "Ekiden")
	memberID := saveTestMember(t, client, testTeamID, "Taro", nil)

	for _, invited := range []shared.RaceID{raceID, otherRaceID} {
		invitation, err := participation.InvitationFactory{}.New(invited, memberID, false)
		if err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, invitation); err != nil {
			t.Fatalf("Save: %v", err)
		}
	}
	if err := repository.Save(ctx, participation.InvitationFactory{}.Reconstruct(raceID, memberID)); err != nil {
		t.Errorf("Save of a stored invitation error = %v, want nil", err)
	}

	found, err := repository.Find(ctx, raceID, memberID)
	if err != nil || found.RaceID() != raceID || found.MemberID() != memberID {
		t.Errorf("Find = %+v, %v, want the invitation of %s to %s", found, err, memberID, raceID)
	}
	listed, err := repository.ListByMember(ctx, memberID)
	if err != nil || len(listed) != 2 || listed[0].RaceID() != raceID || listed[1].RaceID() != otherRaceID {
		t.Errorf("ListByMember = %+v, %v, want the invitations to %s and %s", listed, err, raceID, otherRaceID)
	}

	t.Run("deletes the answer and the entry membership with the invitation", func(t *testing.T) {
		mustExec(t, client, "INSERT INTO attendances (race_id, roster_member_id, status) VALUES ($1, $2, 'attending')", string(raceID), string(memberID))
		entryID := random.IDGenerator{}.NewID()
		mustExec(t, client, "INSERT INTO entries (id, race_id, name) VALUES ($1, $2, 'Team A')", entryID, string(raceID))
		mustExec(t, client, "INSERT INTO entry_members (entry_id, race_id, roster_member_id) VALUES ($1, $2, $3)", entryID, string(raceID), string(memberID))

		if err := repository.Delete(ctx, raceID, memberID); err != nil {
			t.Fatalf("Delete: %v", err)
		}
		if _, err := repository.Find(ctx, raceID, memberID); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Find after Delete error = %v, want ErrNotFound", err)
		}
		if count := countRows(t, client, "SELECT count(*) FROM attendances WHERE race_id = $1", string(raceID)); count != 0 {
			t.Errorf("%d answers left, want 0", count)
		}
		if count := countRows(t, client, "SELECT count(*) FROM entry_members WHERE race_id = $1", string(raceID)); count != 0 {
			t.Errorf("%d entry members left, want 0", count)
		}
		if _, err := repository.Find(ctx, otherRaceID, memberID); err != nil {
			t.Errorf("Find of the other invitation error = %v, want it kept", err)
		}
	})
}
