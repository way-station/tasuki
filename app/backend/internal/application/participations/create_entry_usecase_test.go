package participations

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func TestCreateEntryUseCase(t *testing.T) {
	create := func(races *fakeRaceRepository, entries *fakeEntryRepository, name string) (EntryDTO, error) {
		command := ParticipationCommandFactory{}.CreateEntry("r1", name)
		return NewCreateEntryUseCase(races, entries, participation.NewEntryFactory(fixedIDs{}), ParticipationDTOFactory{}, teamID).Execute(context.Background(), command)
	}

	t.Run("saves a new, empty entry", func(t *testing.T) {
		entries := entriesOf()
		got, err := create(&fakeRaceRepository{stored: teamRace(t, teamID)}, entries, " Team A ")
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}
		if got.ID() != "e1" || got.Name() != "Team A" || len(got.Members()) != 0 {
			t.Errorf("returned %+v, want an empty entry e1 Team A", got)
		}
		if len(entries.saved) != 1 || entries.saved[0].RaceID() != "r1" || entries.saved[0].Name() != "Team A" {
			t.Errorf("saved %+v, want Team A of r1", entries.saved)
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		entries := entriesOf()
		if _, err := create(&fakeRaceRepository{stored: teamRace(t, teamID)}, entries, " "); !isValidation(err) {
			t.Errorf("Execute() error = %v, want a validation error", err)
		}
		if len(entries.saved) != 0 {
			t.Error("an entry with a blank name was saved")
		}
	})

	t.Run("rejects a second entry", func(t *testing.T) {
		entries := entriesOf(entry("r1"))
		if _, err := create(&fakeRaceRepository{stored: teamRace(t, teamID)}, entries, "Team B"); !isState(err) {
			t.Errorf("Execute() error = %v, want a state error", err)
		}
		if len(entries.saved) != 0 {
			t.Error("a second entry was saved")
		}
	})

	t.Run("returns not found for a race of another team", func(t *testing.T) {
		entries := entriesOf()
		if _, err := create(&fakeRaceRepository{stored: teamRace(t, "t2")}, entries, "Team A"); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Execute() error = %v, want %v", err, shared.ErrNotFound)
		}
		if len(entries.saved) != 0 {
			t.Error("an entry of another team's race was saved")
		}
	})

	t.Run("returns the save error", func(t *testing.T) {
		entries := entriesOf()
		entries.saveErr = errDatabase
		if _, err := create(&fakeRaceRepository{stored: teamRace(t, teamID)}, entries, "Team A"); !errors.Is(err, errDatabase) {
			t.Errorf("Execute() error = %v, want %v", err, errDatabase)
		}
	})
}
