//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

const rosterTestTeamID shared.TeamID = "01900000-0000-7000-8000-000000000001"

func TestRosterMemberRepository(t *testing.T) {
	client := newTestClient(t)
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Team')", string(rosterTestTeamID))
	ctx := context.Background()
	factory := roster.NewMemberFactory(random.IDGenerator{})
	repository := NewRosterMemberRepository(client, factory)

	created, err := factory.New(rosterTestTeamID, "Taro")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, created); err != nil {
		t.Fatalf("Save new: %v", err)
	}
	found, err := repository.FindByID(ctx, created.ID())
	if err != nil || found.TeamID() != rosterTestTeamID || found.Name() != "Taro" || found.Left() || found.Average() != roster.DefaultAverage() || found.LapCount() != 0 {
		t.Fatalf("FindByID = %+v, %v, want Taro of the team who has not left, with the default average", found, err)
	}

	t.Run("stores the average time and the lap count", func(t *testing.T) {
		lapTime, err := shared.ParsePace("3:30")
		if err != nil {
			t.Fatal(err)
		}
		roster.AverageService{}.Recalculate(found, []shared.Pace{lapTime, lapTime})
		if err := repository.Save(ctx, found); err != nil {
			t.Fatalf("Save average: %v", err)
		}
		stored, err := repository.FindByID(ctx, created.ID())
		if err != nil || stored.Average().String() != "3:30" || stored.LapCount() != 2 {
			t.Errorf("FindByID = %+v, %v, want 3:30 over 2 laps", stored, err)
		}
	})

	t.Run("stores leaving and rejoining, and keeps the average time", func(t *testing.T) {
		found.Leave(testNow)
		if err := repository.Save(ctx, found); err != nil {
			t.Fatalf("Save leave: %v", err)
		}
		left, err := repository.FindByID(ctx, created.ID())
		if err != nil {
			t.Fatal(err)
		}
		if at, ok := left.LeftAt(); !ok || !at.Equal(testNow) {
			t.Errorf("LeftAt = %v, %v, want %v", at, ok, testNow)
		}

		left.Rejoin()
		if err := repository.Save(ctx, left); err != nil {
			t.Fatalf("Save rejoin: %v", err)
		}
		if rejoined, err := repository.FindByID(ctx, created.ID()); err != nil || rejoined.Left() {
			t.Errorf("FindByID after rejoining = %+v, %v, want not left", rejoined, err)
		}
		if count := countRows(t, client, "SELECT count(*) FROM roster_members WHERE id = $1 AND average_seconds = 210 AND lap_count = 2", string(created.ID())); count != 1 {
			t.Error("the average time or the lap count was changed")
		}
	})

	if _, err := repository.FindByID(ctx, shared.MemberID(random.IDGenerator{}.NewID())); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindByID(unknown) error = %v, want ErrNotFound", err)
	}
}
