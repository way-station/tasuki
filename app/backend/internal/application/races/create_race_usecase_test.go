package races

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

const teamID shared.TeamID = "t1"

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "r1" }

type fakeRaceRepository struct {
	stored  *race.Race
	saved   *race.Race
	saveErr error
}

func (fake *fakeRaceRepository) FindByID(_ context.Context, id shared.RaceID) (*race.Race, error) {
	if fake.stored == nil || fake.stored.ID() != id {
		return nil, fmt.Errorf("race %s: %w", id, shared.ErrNotFound)
	}
	return fake.stored, nil
}

func (fake *fakeRaceRepository) Save(_ context.Context, saved *race.Race) error {
	fake.saved = saved
	return fake.saveErr
}

func TestCreateRaceUseCase(t *testing.T) {
	command, err := RaceCommandFactory{}.CreateRace("Ekiden", distanceInput())
	if err != nil {
		t.Fatal(err)
	}

	t.Run("saves a new race of the team", func(t *testing.T) {
		races := &fakeRaceRepository{}
		id, err := NewCreateRaceUseCase(races, race.NewFactory(fixedIDs{}), teamID).Execute(context.Background(), command)
		if err != nil || id != "r1" {
			t.Fatalf("Execute() = %q, %v, want r1", id, err)
		}
		if races.saved == nil || races.saved.TeamID() != teamID || races.saved.Name() != "Ekiden" || races.saved.Version() != 0 {
			t.Errorf("saved %+v, want a new race of %s", races.saved, teamID)
		}
	})

	t.Run("rejects a blank name", func(t *testing.T) {
		blank, err := RaceCommandFactory{}.CreateRace(" ", distanceInput())
		if err != nil {
			t.Fatal(err)
		}
		races := &fakeRaceRepository{}
		if _, err := NewCreateRaceUseCase(races, race.NewFactory(fixedIDs{}), teamID).Execute(context.Background(), blank); !isValidation(err) {
			t.Errorf("Execute() error = %v, want a validation error", err)
		}
		if races.saved != nil {
			t.Error("a race with a blank name was saved")
		}
	})

	t.Run("returns the save error", func(t *testing.T) {
		saveErr := errors.New("database is down")
		races := &fakeRaceRepository{saveErr: saveErr}
		if _, err := NewCreateRaceUseCase(races, race.NewFactory(fixedIDs{}), teamID).Execute(context.Background(), command); !errors.Is(err, saveErr) {
			t.Errorf("Execute() error = %v, want %v", err, saveErr)
		}
	})
}
