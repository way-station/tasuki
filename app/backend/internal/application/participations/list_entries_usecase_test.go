package participations

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fakeEntryQueryService struct {
	records []EntryRecord
	err     error
	teamID  shared.TeamID
	raceID  shared.RaceID
}

func (fake *fakeEntryQueryService) ListEntries(_ context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]EntryRecord, error) {
	fake.teamID, fake.raceID = teamID, raceID
	return fake.records, fake.err
}

func TestListEntriesUseCase(t *testing.T) {
	t.Run("returns the entries with their members", func(t *testing.T) {
		query := &fakeEntryQueryService{records: []EntryRecord{{ID: "e1", Name: "Team A", Members: []EntryMemberRecord{{MemberID: "m1", Name: "Taro"}, {MemberID: "m2", Name: "Jiro"}}}}}
		got, err := NewListEntriesUseCase(query, ParticipationDTOFactory{}, teamID).Execute(context.Background(), "r1")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if query.teamID != teamID || query.raceID != "r1" {
			t.Errorf("read race %s of team %s, want r1 of %s", query.raceID, query.teamID, teamID)
		}
		if len(got) != 1 || got[0].ID() != "e1" || got[0].Name() != "Team A" {
			t.Fatalf("got %+v, want entry e1 Team A", got)
		}
		members := got[0].Members()
		if len(members) != 2 || members[0].MemberID() != "m1" || members[0].Name() != "Taro" || members[1].MemberID() != "m2" {
			t.Errorf("members = %+v, want Taro and Jiro", members)
		}
	})

	t.Run("returns the error of a race the team does not have", func(t *testing.T) {
		query := &fakeEntryQueryService{err: shared.ErrNotFound}
		if _, err := NewListEntriesUseCase(query, ParticipationDTOFactory{}, teamID).Execute(context.Background(), "r1"); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Execute() error = %v, want %v", err, shared.ErrNotFound)
		}
	})
}
