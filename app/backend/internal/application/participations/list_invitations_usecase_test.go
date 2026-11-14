package participations

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fakeInvitationQueryService struct {
	records []InvitationRecord
	err     error
	teamID  shared.TeamID
	raceID  shared.RaceID
}

func (fake *fakeInvitationQueryService) ListInvitations(_ context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]InvitationRecord, error) {
	fake.teamID, fake.raceID = teamID, raceID
	return fake.records, fake.err
}

func TestListInvitationsUseCase(t *testing.T) {
	attending := participation.Attending
	records := []InvitationRecord{
		{MemberID: "m1", Name: "Taro", Invited: true, Attendance: &attending},
		{MemberID: "m2", Name: "Jiro", Invited: true},
		{MemberID: "m3", Name: "Saburo", Invited: false},
	}
	list := func(t *testing.T, caller auth.MeDTO) []InvitationDTO {
		t.Helper()
		query := &fakeInvitationQueryService{records: records}
		got, err := NewListInvitationsUseCase(query, participation.AttendancePolicy{}, ParticipationDTOFactory{}, teamID).Execute(context.Background(), caller, "r1")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if query.teamID != teamID || query.raceID != "r1" {
			t.Errorf("read race %s of team %s, want r1 of %s", query.raceID, query.teamID, teamID)
		}
		if len(got) != len(records) {
			t.Fatalf("got %d members, want %d", len(got), len(records))
		}
		return got
	}

	t.Run("returns every member with the answers to an organizer", func(t *testing.T) {
		got := list(t, organizer())
		if answer, ok := got[0].Attendance(); !ok || answer != participation.Attending || got[0].MemberID() != "m1" || got[0].Name() != "Taro" || !got[0].Invited() {
			t.Errorf("first member = %+v, want Taro, invited, attending", got[0])
		}
		if _, ok := got[1].Attendance(); ok || !got[1].Invited() {
			t.Errorf("second member = %+v, want invited without an answer", got[1])
		}
		if got[2].Invited() {
			t.Errorf("third member = %+v, want not invited", got[2])
		}
	})

	t.Run("leaves out the answers for anyone else", func(t *testing.T) {
		got := list(t, member())
		if _, ok := got[0].Attendance(); ok || !got[0].Invited() {
			t.Errorf("first member = %+v, want invited without the answer", got[0])
		}
	})

	t.Run("returns the error of a race the team does not have", func(t *testing.T) {
		query := &fakeInvitationQueryService{err: shared.ErrNotFound}
		if _, err := NewListInvitationsUseCase(query, participation.AttendancePolicy{}, ParticipationDTOFactory{}, teamID).Execute(context.Background(), organizer(), "r1"); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Execute() error = %v, want %v", err, shared.ErrNotFound)
		}
	})
}
