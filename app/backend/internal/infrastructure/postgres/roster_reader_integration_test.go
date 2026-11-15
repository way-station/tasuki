//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const (
	readerTeamID   = "01900000-0000-7000-8000-000000000d01"
	readerOtherID  = "01900000-0000-7000-8000-000000000d02"
	readerTaroID   = "01900000-0000-7000-8000-000000000d21"
	readerHanakoID = "01900000-0000-7000-8000-000000000d22"
	readerOtherMID = "01900000-0000-7000-8000-000000000d29"
	readerUserID   = "01900000-0000-7000-8000-000000000d31"
	readerOldRace  = "01900000-0000-7000-8000-000000000d41"
	readerNewRace  = "01900000-0000-7000-8000-000000000d42"
	readerOtherRce = "01900000-0000-7000-8000-000000000d49"
	readerInviteID = "01900000-0000-7000-8000-000000000d51"
	readerUsedID   = "01900000-0000-7000-8000-000000000d52"
	readerLinkID   = "01900000-0000-7000-8000-000000000d61"
)

// saveRosterReaderFixtures saves Taro (left, with a pending link and an old invite link)
// and Hanako (with an open invite link and invitations to two races) of the team, and a
// member and a race of another team.
func saveRosterReaderFixtures(t *testing.T, client *Client) {
	t.Helper()
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Team'), ($2, 'Other')", readerTeamID, readerOtherID)
	mustExec(t, client, `INSERT INTO roster_members (id, team_id, name, left_at, average_seconds, lap_count) VALUES
		($1, $4, 'Taro', now(), 208, 8), ($2, $4, 'Hanako', NULL, 220, 0), ($3, $5, 'Other', NULL, 220, 0)`,
		readerTaroID, readerHanakoID, readerOtherMID, readerTeamID, readerOtherID)
	mustExec(t, client, "INSERT INTO users (id, line_user_id, display_name) VALUES ($1, 'U-taro', 'taro on LINE')", readerUserID)
	mustExec(t, client, `INSERT INTO invite_links (id, roster_member_id, token_hash, expires_at, used_at) VALUES
		($1, $3, 'used', now(), now()), ($2, $4, 'open', '2026-11-20T00:00:00Z', NULL)`, readerUsedID, readerInviteID, readerTaroID, readerHanakoID)
	mustExec(t, client, "INSERT INTO member_links (id, user_id, roster_member_id, invite_link_id, created_at) VALUES ($1, $2, $3, $4, $5)",
		readerLinkID, readerUserID, readerTaroID, readerUsedID, testNow)
	mustExec(t, client, `INSERT INTO races (id, team_id, name, description) VALUES
		($1, $4, 'Old race', '{"race_date":"2026-10-01"}'), ($2, $4, 'New race', '{"race_date":"2026-12-06"}'), ($3, $5, 'Other race', '{"race_date":"2026-12-24"}')`,
		readerOldRace, readerNewRace, readerOtherRce, readerTeamID, readerOtherID)
	mustExec(t, client, "INSERT INTO race_invitations (race_id, roster_member_id) VALUES ($1, $3), ($2, $3)", readerOldRace, readerNewRace, readerHanakoID)
}

func TestRosterReader(t *testing.T) {
	client := newTestClient(t)
	saveRosterReaderFixtures(t, client)
	ctx := context.Background()
	reader := NewRosterReader(client)

	t.Run("lists the team's members by name with their links and open invite link", func(t *testing.T) {
		members, err := reader.ListRosterMembers(ctx, readerTeamID)
		if err != nil || len(members) != 2 {
			t.Fatalf("ListRosterMembers = %+v, %v, want Hanako and Taro", members, err)
		}
		hanako, taro := members[0], members[1]
		if hanako.Name != "Hanako" || hanako.Left || len(hanako.Links) != 0 || hanako.InviteLink == nil || hanako.InviteLink.ID != readerInviteID {
			t.Errorf("members[0] = %+v, want Hanako with the open invite link", hanako)
		}
		if taro.Name != "Taro" || !taro.Left || len(taro.Links) != 1 || taro.Links[0].ID != readerLinkID || taro.Links[0].UserID != readerUserID || taro.InviteLink != nil {
			t.Errorf("members[1] = %+v, want Taro who left, with the pending link and no open invite link", taro)
		}
	})

	t.Run("gets a member with their average", func(t *testing.T) {
		taro, err := reader.GetRosterMember(ctx, readerTeamID, readerTaroID)
		if err != nil || taro.Name != "Taro" || !taro.Left || taro.Average.String() != "3:28" || taro.LapCount != 8 {
			t.Errorf("GetRosterMember = %+v, %v, want Taro 3:28 over 8 laps", taro, err)
		}
		if _, err := reader.GetRosterMember(ctx, readerTeamID, readerOtherMID); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("GetRosterMember of another team error = %v, want ErrNotFound", err)
		}
	})

	t.Run("lists the pending links with the LINE display name", func(t *testing.T) {
		links, err := reader.ListPendingLinks(ctx, readerTeamID)
		if err != nil || len(links) != 1 || links[0].LinkID != readerLinkID || links[0].LineDisplayName != "taro on LINE" || links[0].MemberName != "Taro" || !links[0].CreatedAt.Equal(testNow) {
			t.Errorf("ListPendingLinks = %+v, %v, want the link of Taro", links, err)
		}
		if links, err := reader.ListPendingLinks(ctx, readerOtherID); err != nil || len(links) != 0 {
			t.Errorf("ListPendingLinks of another team = %+v, %v, want none", links, err)
		}
	})

	t.Run("lists the invited races with their dates", func(t *testing.T) {
		races, err := reader.ListInvitedRaces(ctx, readerHanakoID)
		if err != nil || len(races) != 2 {
			t.Fatalf("ListInvitedRaces = %+v, %v, want two races", races, err)
		}
		dates := map[shared.RaceID]string{}
		for _, race := range races {
			dates[race.RaceID] = race.RaceDate.String()
		}
		if dates[readerOldRace] != "2026-10-01" || dates[readerNewRace] != "2026-12-06" {
			t.Errorf("ListInvitedRaces dates = %v", dates)
		}
	})

	t.Run("reads the team's races newest first", func(t *testing.T) {
		races, err := reader.RacesOfTeam(ctx, readerTeamID, []shared.RaceID{readerOldRace, readerNewRace, readerOtherRce})
		if err != nil || len(races) != 2 || races[0].ID != readerNewRace || races[0].Name != "New race" || races[1].ID != readerOldRace {
			t.Errorf("RacesOfTeam = %+v, %v, want the new race and then the old one", races, err)
		}
		if races, err := reader.RacesOfTeam(ctx, readerTeamID, nil); err != nil || len(races) != 0 {
			t.Errorf("RacesOfTeam(none) = %+v, %v, want none", races, err)
		}
	})
}
