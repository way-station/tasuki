//go:build integration

package postgres

import (
	"context"
	"testing"
)

func TestRequireTeam(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	if err := RequireTeam(ctx, client, testTeamID); err == nil {
		t.Error("RequireTeam before the row exists: expected error, got nil")
	}
	saveTestTeam(t, client)
	if err := RequireTeam(ctx, client, testTeamID); err != nil {
		t.Errorf("RequireTeam after the row exists: %v", err)
	}
}
