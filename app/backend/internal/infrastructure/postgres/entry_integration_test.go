//go:build integration

package postgres

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

func TestEntryRepository(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	ctx := context.Background()
	factory := participation.NewEntryFactory(random.IDGenerator{})
	repository := NewEntryRepository(client, factory)
	raceID := saveTestRace(t, client, testTeamID, "Ekiden")
	invitations := NewInvitationRepository(client, participation.InvitationFactory{})
	var members []shared.MemberID
	for _, name := range []string{"Taro", "Jiro", "Saburo"} {
		memberID := saveTestMember(t, client, testTeamID, name, nil)
		if err := invitations.Save(ctx, participation.InvitationFactory{}.Reconstruct(raceID, memberID)); err != nil {
			t.Fatal(err)
		}
		members = append(members, memberID)
	}

	entry, err := factory.New(raceID, "Team A", 0)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, entry); err != nil {
		t.Fatalf("Save new: %v", err)
	}
	found, err := repository.FindByID(ctx, entry.ID())
	if err != nil || found.RaceID() != raceID || found.Name() != "Team A" || len(found.Members()) != 0 {
		t.Fatalf("FindByID = %+v, %v, want an empty Team A of the race", found, err)
	}

	t.Run("stores the members in the order they were added", func(t *testing.T) {
		for _, memberID := range members {
			if err := found.Add(memberID, participation.Attending); err != nil {
				t.Fatal(err)
			}
		}
		if err := repository.Save(ctx, found); err != nil {
			t.Fatalf("Save members: %v", err)
		}
		listed, err := repository.ListByRace(ctx, raceID)
		if err != nil || len(listed) != 1 || listed[0].ID() != entry.ID() || !slices.Equal(listed[0].Members(), members) {
			t.Errorf("ListByRace = %+v, %v, want the entry with %v", listed, err, members)
		}
	})

	t.Run("removes a member and empties their legs", func(t *testing.T) {
		runningOrderID := random.IDGenerator{}.NewID()
		mustExec(t, client, "INSERT INTO running_orders (id, entry_id, status) VALUES ($1, $2, 'draft')", runningOrderID, string(entry.ID()))
		mustExec(t, client, "INSERT INTO running_order_legs (running_order_id, entry_id, leg_number, roster_member_id) VALUES ($1, $2, 1, $3)", runningOrderID, string(entry.ID()), string(members[1]))

		current, err := repository.FindByID(ctx, entry.ID())
		if err != nil {
			t.Fatal(err)
		}
		current.Remove(members[1])
		if err := repository.Save(ctx, current); err != nil {
			t.Fatalf("Save removal: %v", err)
		}
		if saved, err := repository.FindByID(ctx, entry.ID()); err != nil || !slices.Equal(saved.Members(), []shared.MemberID{members[0], members[2]}) {
			t.Errorf("FindByID after the removal = %+v, %v, want the first and third member", saved, err)
		}
		if count := countRows(t, client, "SELECT count(*) FROM running_order_legs WHERE running_order_id = $1 AND roster_member_id IS NULL", runningOrderID); count != 1 {
			t.Errorf("%d empty legs, want 1", count)
		}
	})

	t.Run("joins the unit of work and rolls back with it", func(t *testing.T) {
		current, err := repository.FindByID(ctx, entry.ID())
		if err != nil {
			t.Fatal(err)
		}
		current.Remove(members[0])
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
		if saved, err := repository.FindByID(ctx, entry.ID()); err != nil || !saved.Has(members[0]) {
			t.Errorf("FindByID after the rollback = %+v, %v, want the first member kept", saved, err)
		}
	})

	if _, err := repository.FindByID(ctx, shared.EntryID(random.IDGenerator{}.NewID())); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("FindByID(unknown) error = %v, want ErrNotFound", err)
	}
}
