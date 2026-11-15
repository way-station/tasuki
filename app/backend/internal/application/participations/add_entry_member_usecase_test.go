package participations

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestAddEntryMemberUseCase(t *testing.T) {
	type answers = map[shared.MemberID]participation.AttendanceStatus
	add := func(races *fakeRaceRepository, entries *fakeEntryRepository, attendances *fakeAttendanceRepository) error {
		command := ParticipationCommandFactory{}.EntryMember("r1", "e1", "m1")
		return NewAddEntryMemberUseCase(races, entries, attendances, teamID).Execute(context.Background(), command)
	}

	t.Run("adds an attending member", func(t *testing.T) {
		entries := entriesOf(entry("r1", "m2"))
		if err := add(&fakeRaceRepository{stored: teamRace(t, teamID)}, entries, answered("r1", answers{"m1": participation.Attending})); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(entries.saved) != 1 || !entries.saved[0].Has("m1") || !entries.saved[0].Has("m2") {
			t.Errorf("saved %+v, want the entry with m1 and m2", entries.saved)
		}
	})

	rejected := []struct {
		name   string
		before answers
	}{
		{name: "not invited or not answered", before: answers{}},
		{name: "undecided", before: answers{"m1": participation.Undecided}},
		{name: "not attending", before: answers{"m1": participation.NotAttending}},
	}
	for _, testCase := range rejected {
		t.Run("rejects a member who is "+testCase.name, func(t *testing.T) {
			entries := entriesOf(entry("r1"))
			if err := add(&fakeRaceRepository{stored: teamRace(t, teamID)}, entries, answered("r1", testCase.before)); !isValidation(err) {
				t.Errorf("Execute() error = %v, want a validation error", err)
			}
			if len(entries.saved) != 0 {
				t.Error("the member was added")
			}
		})
	}

	failures := []struct {
		name        string
		races       *fakeRaceRepository
		entryRace   shared.RaceID
		attendances *fakeAttendanceRepository
		want        error
	}{
		{name: "the race belongs to another team", races: &fakeRaceRepository{stored: teamRace(t, "t2")}, entryRace: "r1", attendances: answered("r1", answers{}), want: shared.ErrNotFound},
		{name: "the entry belongs to another race", races: &fakeRaceRepository{stored: teamRace(t, teamID)}, entryRace: "r2", attendances: answered("r1", answers{}), want: shared.ErrNotFound},
		{name: "the answer cannot be read", races: &fakeRaceRepository{stored: teamRace(t, teamID)}, entryRace: "r1", attendances: &fakeAttendanceRepository{findErr: errDatabase}, want: errDatabase},
	}
	for _, testCase := range failures {
		t.Run(testCase.name, func(t *testing.T) {
			entries := entriesOf(entry(testCase.entryRace))
			if err := add(testCase.races, entries, testCase.attendances); !errors.Is(err, testCase.want) {
				t.Errorf("Execute() error = %v, want %v", err, testCase.want)
			}
			if len(entries.saved) != 0 {
				t.Error("the member was added")
			}
		})
	}
}
