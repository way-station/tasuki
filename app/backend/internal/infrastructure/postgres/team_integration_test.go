//go:build integration

package postgres

import (
	"context"
	"testing"
)

func TestRequireTeam(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	if err := RequireTeam(ctx, client, rosterTestTeamID); err == nil {
		t.Error("RequireTeam before the row exists: expected error, got nil")
	}
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Team')", string(rosterTestTeamID))
	if err := RequireTeam(ctx, client, rosterTestTeamID); err != nil {
		t.Errorf("RequireTeam after the row exists: %v", err)
	}
}

func TestCreateTeamIfMissing(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	for range 2 {
		if err := CreateTeamIfMissing(ctx, client, rosterTestTeamID, "Development team"); err != nil {
			t.Fatalf("CreateTeamIfMissing: %v", err)
		}
	}
	if err := RequireTeam(ctx, client, rosterTestTeamID); err != nil {
		t.Errorf("RequireTeam after CreateTeamIfMissing: %v", err)
	}
}
