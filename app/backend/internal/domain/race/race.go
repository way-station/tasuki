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
func (r *Race) ID() shared.RaceID { return r.id }

// TeamID returns the team the race belongs to.
func (r *Race) TeamID() shared.TeamID { return r.teamID }

// Name returns the race name.
func (r *Race) Name() string { return r.name }

// Description returns the race description.
func (r *Race) Description() Description { return r.description }

// Version returns the version used to detect concurrent edits.
func (r *Race) Version() int { return r.version }

// LoadedVersion returns the version the race was read with. Repositories store it only if
// the stored version is still this one.
func (r *Race) LoadedVersion() int { return r.loaded }

// Update replaces the name and the description. It returns shared.ErrVersionConflict when
// the race changed since the caller read version. When the type or the number of legs
// changes, the caller adjusts the running orders.
func (r *Race) Update(name string, d Description, version int) error {
	if version != r.version {
		return shared.ErrVersionConflict
	}
	name, err := validName(name)
	if err != nil {
		return err
	}
	r.name = name
	r.description = d
	r.version++
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
func (f Factory) New(teamID shared.TeamID, name string, d Description) (*Race, error) {
	name, err := validName(name)
	if err != nil {
		return nil, err
	}
	return &Race{id: shared.RaceID(f.ids.NewID()), teamID: teamID, name: name, description: d}, nil
}

// Reconstruct returns a race read from storage.
func (f Factory) Reconstruct(id shared.RaceID, teamID shared.TeamID, name string, d Description, version int) *Race {
	return &Race{id: id, teamID: teamID, name: name, description: d, version: version, loaded: version}
}

// Repository stores races. Errors for missing races wrap shared.ErrNotFound.
type Repository interface {
	FindByID(ctx context.Context, id shared.RaceID) (*Race, error)
	// Save inserts or updates the race. An update stores the race only if the stored
	// version is LoadedVersion, and returns shared.ErrVersionConflict otherwise.
	Save(ctx context.Context, r *Race) error
}
