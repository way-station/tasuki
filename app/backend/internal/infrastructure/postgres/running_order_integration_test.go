//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

// runningOrderFixture is a distance race of two legs of the team, with an entry of
// invited members who answered attending.
type runningOrderFixture struct {
	raceID  shared.RaceID
	entryID shared.EntryID
	members []shared.MemberID
}

// seedRunningOrderRace stores the race and an entry with a member per name, whose average
// times are 3:30, 3:40, and so on.
func seedRunningOrderRace(t *testing.T, client *Client, teamID shared.TeamID, names ...string) runningOrderFixture {
	t.Helper()
	created, err := race.NewFactory(random.IDGenerator{}).New(teamID, "Ekiden", distanceDescription(t, 5, 3.2))
	if err != nil {
		t.Fatal(err)
	}
	if err := NewRaceRepository(client, race.NewFactory(random.IDGenerator{})).Save(context.Background(), created); err != nil {
		t.Fatal(err)
	}
	fixture := runningOrderFixture{raceID: created.ID(), entryID: shared.EntryID(random.IDGenerator{}.NewID())}
	mustExec(t, client, "INSERT INTO entries (id, race_id, name) VALUES ($1, $2, 'Team A')", string(fixture.entryID), string(fixture.raceID))
	for i, name := range names {
		memberID := shared.MemberID(random.IDGenerator{}.NewID())
		mustExec(t, client, "INSERT INTO roster_members (id, team_id, name, average_seconds) VALUES ($1, $2, $3, $4)", string(memberID), string(teamID), name, 210+10*i)
		mustExec(t, client, "INSERT INTO race_invitations (race_id, roster_member_id) VALUES ($1, $2)", string(fixture.raceID), string(memberID))
		mustExec(t, client, "INSERT INTO attendances (race_id, roster_member_id, status) VALUES ($1, $2, 'attending')", string(fixture.raceID), string(memberID))
		mustExec(t, client, "INSERT INTO entry_members (entry_id, race_id, roster_member_id) VALUES ($1, $2, $3)", string(fixture.entryID), string(fixture.raceID), string(memberID))
		fixture.members = append(fixture.members, memberID)
	}
	return fixture
}

func legsOf(runners ...*shared.MemberID) []runningorder.Assignment {
	legs := make([]runningorder.Assignment, 0, len(runners))
	for i, runner := range runners {
		legs = append(legs, runningorder.Assignment{Number: i + 1, Runner: runner})
	}
	return legs
}

func runnersOf(runningOrder *runningorder.RunningOrder) []string {
	var runners []string
	for _, leg := range runningOrder.Legs() {
		runner, _ := leg.Runner()
		runners = append(runners, string(runner))
	}
	return runners
}

func TestRunningOrderRepository(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	ctx := context.Background()
	fixture := seedRunningOrderRace(t, client, testTeamID, "Taro", "Jiro")
	taro, jiro := fixture.members[0], fixture.members[1]
	factory := runningorder.NewFactory(random.IDGenerator{})
	repository := NewRunningOrderRepository(client, factory)
	shape := runningorder.Shape{LegCount: 2}

	if _, err := repository.FindByEntry(ctx, fixture.entryID); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindByEntry before saving error = %v, want ErrNotFound", err)
	}
	created, err := factory.New(fixture.entryID, legsOf(&taro, nil), shape, fixture.members)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, created); err != nil {
		t.Fatalf("Save new: %v", err)
	}
	found, err := repository.FindByEntry(ctx, fixture.entryID)
	if err != nil || found.ID() != created.ID() || found.Status() != runningorder.Draft || found.Version() != 0 || runnersOf(found)[0] != string(taro) || runnersOf(found)[1] != "" {
		t.Fatalf("FindByEntry = %+v, %v, want the draft with Taro on leg 1", found, err)
	}

	t.Run("replaces the legs and confirms when the version is still the one read", func(t *testing.T) {
		stale, err := repository.FindByEntry(ctx, fixture.entryID)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := found.Replace(legsOf(&jiro, &taro), shape, fixture.members, 0); err != nil {
			t.Fatal(err)
		}
		if err := found.Confirm(1); err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, found); err != nil {
			t.Fatalf("Save: %v", err)
		}
		saved, err := repository.FindByEntry(ctx, fixture.entryID)
		if err != nil || saved.Status() != runningorder.Confirmed || saved.Version() != 2 || runnersOf(saved)[0] != string(jiro) || runnersOf(saved)[1] != string(taro) {
			t.Errorf("FindByEntry after Save = %+v, %v, want the confirmed order with Jiro and Taro", saved, err)
		}
		if count := countRunningOrderRows(t, client, "SELECT count(*) FROM running_orders WHERE entry_id = $1 AND confirmed_at IS NOT NULL", string(fixture.entryID)); count != 1 {
			t.Error("confirmed_at was not set")
		}

		if _, err := stale.Replace(legsOf(nil, nil), shape, fixture.members, 0); err != nil {
			t.Fatal(err)
		}
		if err := repository.Save(ctx, stale); !errors.Is(err, shared.ErrVersionConflict) {
			t.Errorf("Save stale error = %v, want ErrVersionConflict", err)
		}
	})

	t.Run("joins the unit of work and rolls back with it", func(t *testing.T) {
		current, err := repository.FindByEntry(ctx, fixture.entryID)
		if err != nil {
			t.Fatal(err)
		}
		current.RemoveRunner(taro)
		rollback := errors.New("roll back")
		err = NewUnitOfWork(client).Run(ctx, func(ctx context.Context) error {
			if err := repository.Save(ctx, current); err != nil {
				return err
			}
			return rollback
		})
		if !errors.Is(err, rollback) {
			t.Fatalf("Run error = %v, want %v", err, rollback)
		}
		if saved, err := repository.FindByEntry(ctx, fixture.entryID); err != nil || runnersOf(saved)[1] != string(taro) {
			t.Errorf("FindByEntry after the rollback = %+v, %v, want Taro kept", saved, err)
		}
	})
}

func countRunningOrderRows(t *testing.T, client *Client, sql string, args ...any) int {
	t.Helper()
	var count int
	if err := client.Pool().QueryRow(context.Background(), sql, args...).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
