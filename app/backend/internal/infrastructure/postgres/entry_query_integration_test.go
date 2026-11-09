//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/participations"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

func TestEntryQueryService(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	const otherTeamID shared.TeamID = "01900000-0000-7000-8000-000000000002"
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Other team')", string(otherTeamID))
	ctx := context.Background()
	raceID := saveTestRace(t, client, testTeamID)
	invitations := NewInvitationRepository(client, participation.InvitationFactory{})
	factory := participation.NewEntryFactory(random.IDGenerator{})
	entries := NewEntryRepository(client, factory)

	queryService := NewEntryQueryService(client)
	if records, err := queryService.ListEntries(ctx, testTeamID, raceID); err != nil || len(records) != 0 {
		t.Errorf("ListEntries before any entry = %+v, %v, want none", records, err)
	}

	entry, err := factory.New(raceID, "Team A", 0)
	if err != nil {
		t.Fatal(err)
	}
	var wantMembers []participations.EntryMemberRecord
	for _, name := range []string{"Taro", "Jiro"} {
		memberID := saveTestMember(t, client, testTeamID, name, nil)
		if err := invitations.Save(ctx, participation.InvitationFactory{}.Reconstruct(raceID, memberID)); err != nil {
			t.Fatal(err)
		}
		if err := entry.Add(memberID, participation.Attending); err != nil {
			t.Fatal(err)
		}
		wantMembers = append(wantMembers, participations.EntryMemberRecord{MemberID: memberID, Name: name})
	}
	if err := entries.Save(ctx, entry); err != nil {
		t.Fatal(err)
	}

	records, err := queryService.ListEntries(ctx, testTeamID, raceID)
	if err != nil {
		t.Fatalf("ListEntries: %v", err)
	}
	if len(records) != 1 || records[0].ID != entry.ID() || records[0].Name != "Team A" || len(records[0].Members) != len(wantMembers) {
		t.Fatalf("ListEntries = %+v, want Team A with %d members", records, len(wantMembers))
	}
	for i, member := range records[0].Members {
		if member != wantMembers[i] {
			t.Errorf("member %d = %+v, want %+v", i, member, wantMembers[i])
		}
	}

	otherRaceID := saveTestRace(t, client, otherTeamID)
	if _, err := queryService.ListEntries(ctx, testTeamID, otherRaceID); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("ListEntries(race of another team) error = %v, want ErrNotFound", err)
	}
}
