package race

import (
	"context"
	"strings"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Race is a race the team enters. It is created by Factory. Races are not deleted.
type Race struct {
	id          shared.RaceID
	teamID      shared.TeamID
	name        string
	description Description
	version     int
	loaded      int // the version it was read with; 0 for a new race
}

// ID returns the race's ID.
func (race *Race) ID() shared.RaceID { return race.id }

// TeamID returns the team the race belongs to.
func (race *Race) TeamID() shared.TeamID { return race.teamID }

// Name returns the race name.
func (race *Race) Name() string { return race.name }

// Description returns the race description.
func (race *Race) Description() Description { return race.description }

// Version returns the version used to detect concurrent edits.
func (race *Race) Version() int { return race.version }

// LoadedVersion returns the version the race was read with. Repositories store it only if
// the stored version is still this one.
func (race *Race) LoadedVersion() int { return race.loaded }

// Update replaces the name and the description. It returns shared.ErrVersionConflict when
// the race changed since the caller read version. When the type or the number of legs
// changes, the caller adjusts the running orders.
func (race *Race) Update(name string, description Description, version int) error {
	if version != race.version {
		return shared.ErrVersionConflict
	}
	name, err := validName(name)
	if err != nil {
		return err
	}
	race.name = name
	race.description = description
	race.version++
	return nil
}

func validName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", shared.NewValidationError("name", "is required")
	}
	return name, nil
}

// Factory assembles Race values.
type Factory struct {
	ids shared.IDGenerator
}

// NewFactory returns a Factory that takes new IDs from ids.
func NewFactory(ids shared.IDGenerator) Factory {
	return Factory{ids: ids}
}

// New returns a new race of team.
func (factory Factory) New(teamID shared.TeamID, name string, description Description) (*Race, error) {
	name, err := validName(name)
	if err != nil {
		return nil, err
	}
	return &Race{id: shared.RaceID(factory.ids.NewID()), teamID: teamID, name: name, description: description}, nil
}

// Reconstruct returns a race read from storage.
func (factory Factory) Reconstruct(id shared.RaceID, teamID shared.TeamID, name string, description Description, version int) *Race {
	return &Race{id: id, teamID: teamID, name: name, description: description, version: version, loaded: version}
}

// Repository stores races. Errors for missing races wrap shared.ErrNotFound.
type Repository interface {
	FindByID(ctx context.Context, id shared.RaceID) (*Race, error)
	// Save inserts or updates the race. An update stores the race only if the stored
	// version is LoadedVersion, and returns shared.ErrVersionConflict otherwise.
	Save(ctx context.Context, race *Race) error
}
