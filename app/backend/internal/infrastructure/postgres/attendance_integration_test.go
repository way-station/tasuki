//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestAttendanceRepository(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	ctx := context.Background()
	repository := NewAttendanceRepository(client, participation.AttendanceFactory{})
	raceID := saveTestRace(t, client, testTeamID, "Ekiden")
	memberID := saveTestMember(t, client, testTeamID, "Taro", nil)
	invitation := participation.InvitationFactory{}.Reconstruct(raceID, memberID)
	if err := NewInvitationRepository(client, participation.InvitationFactory{}).Save(ctx, invitation); err != nil {
		t.Fatal(err)
	}

	if _, err := repository.Find(ctx, raceID, memberID); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("Find before answering error = %v, want ErrNotFound", err)
	}

	first, err := participation.AttendanceFactory{}.New(invitation, participation.Attending)
	if err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, first); err != nil {
		t.Fatalf("Save first: %v", err)
	}
	found, err := repository.Find(ctx, raceID, memberID)
	if err != nil || found.RaceID() != raceID || found.MemberID() != memberID || found.Status() != participation.Attending {
		t.Fatalf("Find = %+v, %v, want attending", found, err)
	}

	if _, err := found.Change(participation.Undecided); err != nil {
		t.Fatal(err)
	}
	if err := repository.Save(ctx, found); err != nil {
		t.Fatalf("Save change: %v", err)
	}
	if changed, err := repository.Find(ctx, raceID, memberID); err != nil || changed.Status() != participation.Undecided {
		t.Errorf("Find after the change = %+v, %v, want undecided", changed, err)
	}
}
