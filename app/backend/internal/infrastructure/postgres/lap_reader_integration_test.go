//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const (
	lapTeamID      = "01900000-0000-7000-8000-000000000a01"
	lapOtherTeamID = "01900000-0000-7000-8000-000000000a02"
	lapRaceID      = "01900000-0000-7000-8000-000000000a11"
	lapEmptyRaceID = "01900000-0000-7000-8000-000000000a12"
	lapTaroID      = "01900000-0000-7000-8000-000000000a21"
	lapJiroID      = "01900000-0000-7000-8000-000000000a22"
	lapHanakoID    = "01900000-0000-7000-8000-000000000a23"
	lapConfirmedID = "01900000-0000-7000-8000-000000000a31"
	lapDraftID     = "01900000-0000-7000-8000-000000000a32"
)

// saveLapRunningOrders saves a race whose confirmed running order has Taro on legs 1 and 3
// and Hanako on leg 2, and whose draft has Jiro, and a race without running orders.
func saveLapRunningOrders(t *testing.T, client *Client) {
	t.Helper()
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Team'), ($2, 'Other')", lapTeamID, lapOtherTeamID)
	mustExec(t, client, "INSERT INTO roster_members (id, team_id, name) VALUES ($1, $4, 'Taro'), ($2, $4, 'Jiro'), ($3, $4, 'Hanako')",
		lapTaroID, lapJiroID, lapHanakoID, lapTeamID)
	mustExec(t, client, "INSERT INTO races (id, team_id, name, description) VALUES ($1, $3, 'Hakone', '{}'), ($2, $3, 'Empty', '{}')",
		lapRaceID, lapEmptyRaceID, lapTeamID)
	mustExec(t, client, "INSERT INTO race_invitations (race_id, roster_member_id) VALUES ($1, $2), ($1, $3), ($1, $4)",
		lapRaceID, lapTaroID, lapJiroID, lapHanakoID)
	mustExec(t, client, "INSERT INTO entries (id, race_id, name) VALUES ($1, $3, 'Team A'), ($2, $3, 'Team B')", lapConfirmedID, lapDraftID, lapRaceID)
	mustExec(t, client, "INSERT INTO entry_members (entry_id, race_id, roster_member_id) VALUES ($1, $3, $4), ($1, $3, $5), ($2, $3, $6)",
		lapConfirmedID, lapDraftID, lapRaceID, lapTaroID, lapHanakoID, lapJiroID)
	mustExec(t, client, "INSERT INTO running_orders (id, entry_id, status) VALUES ($1, $1, 'confirmed'), ($2, $2, 'draft')", lapConfirmedID, lapDraftID)
	mustExec(t, client, `INSERT INTO running_order_legs (running_order_id, entry_id, leg_number, roster_member_id) VALUES
		($1, $1, 1, $3), ($1, $1, 2, $4), ($1, $1, 3, $3), ($2, $2, 1, $5)`, lapConfirmedID, lapDraftID, lapTaroID, lapHanakoID, lapJiroID)
}

func TestLapReader(t *testing.T) {
	client := newTestClient(t)
	saveLapRunningOrders(t, client)
	ctx := context.Background()
	reader := NewLapReader(client)

	t.Run("tells the runners of the confirmed running orders", func(t *testing.T) {
		tests := []struct {
			name     string
			teamID   shared.TeamID
			raceID   shared.RaceID
			memberID shared.MemberID
			want     bool
		}{
			{name: "a runner of two legs", teamID: lapTeamID, raceID: lapRaceID, memberID: lapTaroID, want: true},
			{name: "a runner of a draft", teamID: lapTeamID, raceID: lapRaceID, memberID: lapJiroID, want: false},
			{name: "a runner asked for another race", teamID: lapTeamID, raceID: lapEmptyRaceID, memberID: lapTaroID, want: false},
			{name: "a runner asked for another team", teamID: lapOtherTeamID, raceID: lapRaceID, memberID: lapTaroID, want: false},
		}
		for _, testCase := range tests {
			got, err := reader.IsConfirmedRunner(ctx, testCase.teamID, testCase.raceID, testCase.memberID)
			if err != nil || got != testCase.want {
				t.Errorf("%s: IsConfirmedRunner = %v, %v, want %v", testCase.name, got, err, testCase.want)
			}
		}
	})

	t.Run("lists the confirmed runners once each, by name", func(t *testing.T) {
		runners, err := reader.ConfirmedRunners(ctx, lapTeamID, lapRaceID)
		if err != nil {
			t.Fatalf("ConfirmedRunners: %v", err)
		}
		if len(runners) != 2 || runners[0] != (Runner{MemberID: lapHanakoID, Name: "Hanako"}) || runners[1] != (Runner{MemberID: lapTaroID, Name: "Taro"}) {
			t.Errorf("ConfirmedRunners = %+v, want Hanako and Taro", runners)
		}
	})

	t.Run("lists no runners of a race without running orders", func(t *testing.T) {
		if runners, err := reader.ConfirmedRunners(ctx, lapTeamID, lapEmptyRaceID); err != nil || len(runners) != 0 {
			t.Errorf("ConfirmedRunners = %+v, %v, want none", runners, err)
		}
	})

	t.Run("rejects a race of another team", func(t *testing.T) {
		if _, err := reader.ConfirmedRunners(ctx, lapOtherTeamID, lapRaceID); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("ConfirmedRunners error = %v, want ErrNotFound", err)
		}
	})
}
