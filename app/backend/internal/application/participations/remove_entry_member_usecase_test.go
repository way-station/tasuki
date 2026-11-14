package participations

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestRemoveEntryMemberUseCase(t *testing.T) {
	remove := func(t *testing.T, unitOfWork *fakeUnitOfWork, races *fakeRaceRepository, entries *fakeEntryRepository, runners *fakeRunnerRemover) error {
		t.Helper()
		command := ParticipationCommandFactory{}.EntryMember("r1", "e1", "m1")
		return NewRemoveEntryMemberUseCase(unitOfWork, races, entries, runners, teamID).Execute(context.Background(), command)
	}

	t.Run("removes the member from the entry and the running orders", func(t *testing.T) {
		entries, runners := entriesOf(entry("r1", "m1", "m2")), &fakeRunnerRemover{}
		if err := remove(t, &fakeUnitOfWork{}, &fakeRaceRepository{stored: teamRace(t, teamID)}, entries, runners); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(entries.saved) != 1 || entries.saved[0].Has("m1") || !entries.saved[0].Has("m2") || !entries.savedInTransaction {
			t.Errorf("saved %+v, want the entry without m1 in the transaction", entries.saved)
		}
		if len(runners.removed) != 1 || runners.removed[0] != (removedRunner{raceID: "r1", memberID: "m1", inTransaction: true}) {
			t.Errorf("removed runners %+v, want m1 of r1 in the transaction", runners.removed)
		}
	})

	t.Run("does nothing for a member who is not in the entry", func(t *testing.T) {
		entries, runners := entriesOf(entry("r1", "m2")), &fakeRunnerRemover{}
		if err := remove(t, &fakeUnitOfWork{}, &fakeRaceRepository{stored: teamRace(t, teamID)}, entries, runners); err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if len(entries.saved) != 0 || len(runners.removed) != 0 {
			t.Errorf("saved entries %+v, removed runners %+v, want neither", entries.saved, runners.removed)
		}
	})

	t.Run("rolls back when the running orders cannot be changed", func(t *testing.T) {
		unitOfWork := &fakeUnitOfWork{}
		err := remove(t, unitOfWork, &fakeRaceRepository{stored: teamRace(t, teamID)}, entriesOf(entry("r1", "m1")), &fakeRunnerRemover{err: errDatabase})
		if !errors.Is(err, errDatabase) || !errors.Is(unitOfWork.err, errDatabase) {
			t.Errorf("Execute() error = %v, unit of work got %v, want %v", err, unitOfWork.err, errDatabase)
		}
	})

	failures := []struct {
		name    string
		races   *fakeRaceRepository
		entries *fakeEntryRepository
	}{
		{name: "the race belongs to another team", races: &fakeRaceRepository{stored: teamRace(t, "t2")}, entries: entriesOf(entry("r1", "m1"))},
		{name: "the entry does not exist", races: &fakeRaceRepository{stored: teamRace(t, teamID)}, entries: entriesOf()},
		{name: "the entry belongs to another race", races: &fakeRaceRepository{stored: teamRace(t, teamID)}, entries: entriesOf(entry("r2", "m1"))},
	}
	for _, testCase := range failures {
		t.Run("returns not found when "+testCase.name, func(t *testing.T) {
			runners := &fakeRunnerRemover{}
			if err := remove(t, &fakeUnitOfWork{}, testCase.races, testCase.entries, runners); !errors.Is(err, shared.ErrNotFound) {
				t.Errorf("Execute() error = %v, want %v", err, shared.ErrNotFound)
			}
			if len(testCase.entries.saved) != 0 || len(runners.removed) != 0 {
				t.Error("the entry or the running orders were changed")
			}
		})
	}
}
