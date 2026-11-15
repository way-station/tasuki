package participations

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// CreateEntryUseCase creates an entry in a race (POST /races/{race_id}/entries). Only
// organizers may call it; the middleware checks the role.
type CreateEntryUseCase struct {
	races      race.Repository
	entries    participation.EntryRepository
	factory    participation.EntryFactory
	dtoFactory ParticipationDTOFactory
	teamID     shared.TeamID
}

// NewCreateEntryUseCase returns a CreateEntryUseCase for the team.
func NewCreateEntryUseCase(races race.Repository, entries participation.EntryRepository, factory participation.EntryFactory, dtoFactory ParticipationDTOFactory, teamID shared.TeamID) *CreateEntryUseCase {
	return &CreateEntryUseCase{races: races, entries: entries, factory: factory, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute saves a new, empty entry and returns it. A blank name is a
// shared.ValidationError, and a race that already has an entry is a shared.StateError. A
// race of another team is shared.ErrNotFound.
func (useCase *CreateEntryUseCase) Execute(ctx context.Context, command CreateEntryCommand) (EntryDTO, error) {
	if err := requireTeamRace(ctx, useCase.races, useCase.teamID, command.raceID); err != nil {
		return EntryDTO{}, err
	}
	existing, err := useCase.entries.ListByRace(ctx, command.raceID)
	if err != nil {
		return EntryDTO{}, err
	}
	entry, err := useCase.factory.New(command.raceID, command.name, len(existing))
	if err != nil {
		return EntryDTO{}, err
	}
	if err := useCase.entries.Save(ctx, entry); err != nil {
		return EntryDTO{}, err
	}
	return useCase.dtoFactory.CreateEntry(entry.ID(), entry.Name(), nil), nil
}
