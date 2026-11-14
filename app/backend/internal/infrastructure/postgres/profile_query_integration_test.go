//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestProfileQueryService(t *testing.T) {
	client := newTestClient(t)
	const (
		teamID  = "01900000-0000-7000-8000-000000000b01"
		taroID  = "01900000-0000-7000-8000-000000000b21"
		jiroID  = "01900000-0000-7000-8000-000000000b22"
		unknown = "01900000-0000-7000-8000-000000000b29"
	)
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Team')", teamID)
	mustExec(t, client, "INSERT INTO roster_members (id, team_id, name, average_seconds, lap_count) VALUES ($1, $2, 'Taro', 208, 8)", taroID, teamID)
	mustExec(t, client, "INSERT INTO roster_members (id, team_id, name) VALUES ($1, $2, 'Jiro')", jiroID, teamID)
	service := NewProfileQueryService(client)
	ctx := context.Background()

	profile, err := service.GetProfile(ctx, taroID)
	if err != nil || profile.Name != "Taro" || profile.Average.String() != "3:28" || profile.LapCount != 8 {
		t.Errorf("GetProfile(Taro) = %+v, %v, want Taro 3:28 over 8 laps", profile, err)
	}
	profile, err = service.GetProfile(ctx, jiroID)
	if err != nil || profile.Name != "Jiro" || profile.Average.String() != "3:40" || profile.LapCount != 0 {
		t.Errorf("GetProfile(Jiro) = %+v, %v, want the default 3:40 over 0 laps", profile, err)
	}
	if _, err := service.GetProfile(ctx, shared.MemberID(unknown)); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("GetProfile(unknown) error = %v, want ErrNotFound", err)
	}
}
