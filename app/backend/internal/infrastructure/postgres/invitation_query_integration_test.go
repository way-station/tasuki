//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestInvitationQueryService(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	const otherTeamID shared.TeamID = "01900000-0000-7000-8000-000000000002"
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Other team')", string(otherTeamID))
	ctx := context.Background()
	raceID := saveTestRace(t, client, testTeamID)
	invitations := NewInvitationRepository(client, participation.InvitationFactory{})
	attendances := NewAttendanceRepository(client, participation.AttendanceFactory{})

	attending := saveTestMember(t, client, testTeamID, "Taro", nil)
	unanswered := saveTestMember(t, client, testTeamID, "Jiro", nil)
	notInvited := saveTestMember(t, client, testTeamID, "Saburo", nil)
	saveTestMember(t, client, testTeamID, "Shiro", &testNow)
	saveTestMember(t, client, otherTeamID, "Kuro", nil)
	for _, memberID := range []shared.MemberID{attending, unanswered} {
		if err := invitations.Save(ctx, participation.InvitationFactory{}.Reconstruct(raceID, memberID)); err != nil {
			t.Fatal(err)
		}
	}
	if err := attendances.Save(ctx, participation.AttendanceFactory{}.Reconstruct(raceID, attending, participation.Attending)); err != nil {
		t.Fatal(err)
	}

	queryService := NewInvitationQueryService(client)
	records, err := queryService.ListInvitations(ctx, testTeamID, raceID)
	if err != nil {
		t.Fatalf("ListInvitations: %v", err)
	}
	want := []struct {
		memberID shared.MemberID
		name     string
		invited  bool
		answer   participation.AttendanceStatus
	}{
		{memberID: attending, name: "Taro", invited: true, answer: participation.Attending},
		{memberID: unanswered, name: "Jiro", invited: true},
		{memberID: notInvited, name: "Saburo"},
	}
	if len(records) != len(want) {
		t.Fatalf("ListInvitations = %+v, want %d members who have not left", records, len(want))
	}
	for i, record := range records {
		var answer participation.AttendanceStatus
		if record.Attendance != nil {
			answer = *record.Attendance
		}
		if record.MemberID != want[i].memberID || record.Name != want[i].name || record.Invited != want[i].invited || answer != want[i].answer {
			t.Errorf("member %d = %+v (answer %q), want %+v", i, record, answer, want[i])
		}
	}

	otherRaceID := saveTestRace(t, client, otherTeamID)
	if _, err := queryService.ListInvitations(ctx, testTeamID, otherRaceID); !errors.Is(err, shared.ErrNotFound) {
		t.Errorf("ListInvitations(race of another team) error = %v, want ErrNotFound", err)
	}
}
