package query

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/rosters"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres"
)

var errDatabase = errors.New("database is down")

// fakeRosterReader knows member m1 of team t1 and the races r1 (newest) and r2.
type fakeRosterReader struct {
	askedRaces []shared.RaceID
}

func (fake *fakeRosterReader) ListRosterMembers(context.Context, shared.TeamID) ([]rosters.RosterMemberRecord, error) {
	return []rosters.RosterMemberRecord{{MemberID: "m1", Name: "Taro"}}, nil
}

func (fake *fakeRosterReader) GetRosterMember(_ context.Context, teamID shared.TeamID, memberID shared.MemberID) (rosters.RosterMemberDetailRecord, error) {
	if teamID != "t1" || memberID != "m1" {
		return rosters.RosterMemberDetailRecord{}, fmt.Errorf("roster member %s: %w", memberID, shared.ErrNotFound)
	}
	return rosters.RosterMemberDetailRecord{MemberID: "m1", Name: "Taro", LapCount: 3}, nil
}

func (fake *fakeRosterReader) ListPendingLinks(context.Context, shared.TeamID) ([]rosters.PendingLinkRecord, error) {
	return []rosters.PendingLinkRecord{{LinkID: "k1"}}, nil
}

func (fake *fakeRosterReader) ListInvitedRaces(context.Context, shared.MemberID) ([]roster.InvitedRace, error) {
	return []roster.InvitedRace{{RaceID: "r1"}}, nil
}

func (fake *fakeRosterReader) RacesOfTeam(_ context.Context, _ shared.TeamID, raceIDs []shared.RaceID) ([]postgres.RaceSummary, error) {
	fake.askedRaces = raceIDs
	var races []postgres.RaceSummary
	for _, race := range []postgres.RaceSummary{{ID: "r1", Name: "New race", RaceDate: "2026-12-06"}, {ID: "r2", Name: "Old race", RaceDate: "2026-10-01"}} {
		if slices.Contains(raceIDs, race.ID) {
			races = append(races, race)
		}
	}
	return races, nil
}

type fakeMemberLaps struct {
	byRace map[shared.RaceID][]rosters.LapRecord
	err    error
}

func (fake *fakeMemberLaps) LapsByRace(context.Context, shared.MemberID) (map[shared.RaceID][]rosters.LapRecord, error) {
	return fake.byRace, fake.err
}

func lapOf(t *testing.T, id shared.LapID, seconds int) rosters.LapRecord {
	t.Helper()
	lapTime, err := shared.PaceFromSeconds(seconds)
	if err != nil {
		t.Fatal(err)
	}
	return rosters.LapRecord{ID: id, Time: lapTime}
}

func TestRosterQueryServiceGetRosterMember(t *testing.T) {
	t.Run("puts the laps under the team's races, newest race first", func(t *testing.T) {
		reader := &fakeRosterReader{}
		laps := &fakeMemberLaps{byRace: map[shared.RaceID][]rosters.LapRecord{
			"r2": {lapOf(t, "l1", 200)},
			"r1": {lapOf(t, "l2", 205), lapOf(t, "l3", 210)},
			"r9": {lapOf(t, "l9", 300)}, // a race of another team
		}}
		member, err := NewRosterQueryService(reader, laps).GetRosterMember(context.Background(), "t1", "m1")
		if err != nil {
			t.Fatalf("GetRosterMember() error = %v", err)
		}
		if len(member.LapsByRace) != 2 || member.LapsByRace[0].RaceName != "New race" || len(member.LapsByRace[0].Laps) != 2 || member.LapsByRace[1].RaceID != "r2" || member.LapCount != 3 {
			t.Errorf("GetRosterMember() = %+v, want the new race and then the old one", member)
		}
		if !slices.Equal(reader.askedRaces, []shared.RaceID{"r1", "r2", "r9"}) {
			t.Errorf("asked races = %v, want r1, r2, and r9 in order", reader.askedRaces)
		}
	})

	t.Run("returns no races for a member without laps", func(t *testing.T) {
		member, err := NewRosterQueryService(&fakeRosterReader{}, &fakeMemberLaps{}).GetRosterMember(context.Background(), "t1", "m1")
		if err != nil || member.LapsByRace == nil || len(member.LapsByRace) != 0 {
			t.Errorf("GetRosterMember() = %+v, %v, want an empty list", member, err)
		}
	})

	t.Run("returns the errors of the readers", func(t *testing.T) {
		if _, err := NewRosterQueryService(&fakeRosterReader{}, &fakeMemberLaps{}).GetRosterMember(context.Background(), "t1", "m9"); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("GetRosterMember(m9) error = %v, want shared.ErrNotFound", err)
		}
		if _, err := NewRosterQueryService(&fakeRosterReader{}, &fakeMemberLaps{err: errDatabase}).GetRosterMember(context.Background(), "t1", "m1"); !errors.Is(err, errDatabase) {
			t.Errorf("GetRosterMember() error = %v, want %v", err, errDatabase)
		}
	})
}

func TestRosterQueryServicePassesThrough(t *testing.T) {
	service := NewRosterQueryService(&fakeRosterReader{}, &fakeMemberLaps{})
	ctx := context.Background()
	if members, err := service.ListRosterMembers(ctx, "t1"); err != nil || len(members) != 1 {
		t.Errorf("ListRosterMembers() = %v, %v", members, err)
	}
	if links, err := service.ListPendingLinks(ctx, "t1"); err != nil || len(links) != 1 {
		t.Errorf("ListPendingLinks() = %v, %v", links, err)
	}
	if races, err := service.ListInvitedRaces(ctx, "m1"); err != nil || len(races) != 1 {
		t.Errorf("ListInvitedRaces() = %v, %v", races, err)
	}
}
