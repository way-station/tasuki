package postgres

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// RaceRepository implements race.Repository with PostgreSQL. It runs in the transaction
// of a unit of work when there is one.
type RaceRepository struct {
	queries *sqlcgen.Queries
	factory race.Factory
}

var _ race.Repository = (*RaceRepository)(nil)

// NewRaceRepository returns a RaceRepository that uses client.
func NewRaceRepository(client *Client, factory race.Factory) *RaceRepository {
	return &RaceRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// FindByID returns the race with id.
func (repository *RaceRepository) FindByID(ctx context.Context, id shared.RaceID) (*race.Race, error) {
	raceID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	row, err := queriesFor(ctx, repository.queries).GetRace(ctx, raceID)
	if err != nil {
		return nil, notFound(err, "race")
	}
	params, err := decodeRaceDescription(row.Description)
	if err != nil {
		return nil, err
	}
	description, err := race.NewDescription(params)
	if err != nil {
		return nil, fmt.Errorf("race %s has an invalid stored description: %w", id, err)
	}
	return repository.factory.Reconstruct(id, shared.TeamID(row.TeamID.String()), row.Name, description, int(row.Version)), nil
}

// Save inserts a new race, or updates it if the stored version is still the version it
// was read with. It returns shared.ErrVersionConflict otherwise.
func (repository *RaceRepository) Save(ctx context.Context, saved *race.Race) error {
	raceID, err := parseID(saved.ID())
	if err != nil {
		return err
	}
	teamID, err := parseID(saved.TeamID())
	if err != nil {
		return err
	}
	description, err := encodeRaceDescription(saved.Description())
	if err != nil {
		return err
	}
	version, err := toInt32(saved.Version(), "race version")
	if err != nil {
		return err
	}
	loadedVersion, err := toInt32(saved.LoadedVersion(), "race version")
	if err != nil {
		return err
	}
	affected, err := queriesFor(ctx, repository.queries).SaveRace(ctx, sqlcgen.SaveRaceParams{
		ID:            raceID,
		TeamID:        teamID,
		Name:          saved.Name(),
		Description:   description,
		Version:       version,
		LoadedVersion: loadedVersion,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("race %s: %w", saved.ID(), shared.ErrVersionConflict)
	}
	return nil
}
