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

func TestRosterMemberRepository(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	ctx := context.Background()
	factory := roster.NewMemberFactory(random.IDGenerator{})
	repository := NewRosterMemberRepository(client, factory)

	created, err := factory.New(testTeamID, "Taro")
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, created); err != nil {
		t.Fatalf("Save new: %v", err)
	}
	found, err := repository.FindByID(ctx, created.ID())
	if err != nil || found.TeamID() != testTeamID || found.Name() != "Taro" || found.Left() {
		t.Fatalf("FindByID = %+v, %v, want Taro of the team who has not left", found, err)
	}

	t.Run("stores leaving and rejoining, and keeps the average time", func(t *testing.T) {
		mustExec(t, client, "UPDATE roster_members SET average_seconds = 230, lap_count = 4 WHERE id = $1", string(created.ID()))
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
		if count := countRows(t, client, "SELECT count(*) FROM roster_members WHERE id = $1 AND average_seconds = 230 AND lap_count = 4", string(created.ID())); count != 1 {
			t.Error("the average time or the lap count was changed")
		}
	})

	if _, err := repository.FindByID(ctx, shared.MemberID(random.IDGenerator{}.NewID())); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindByID(unknown) error = %v, want ErrNotFound", err)
	}
}
