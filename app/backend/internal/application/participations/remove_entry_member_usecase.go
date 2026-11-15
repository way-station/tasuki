package participations

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RemoveEntryMemberUseCase removes a roster member from an entry
// (DELETE /races/{race_id}/entries/{entry_id}/members/{roster_member_id}). Only organizers
// may call it; the middleware checks the role.
type RemoveEntryMemberUseCase struct {
	unitOfWork transaction.UnitOfWork
	races      race.Repository
	entries    participation.EntryRepository
	runners    RunnerRemover
	teamID     shared.TeamID
}

// NewRemoveEntryMemberUseCase returns a RemoveEntryMemberUseCase for the team.
func NewRemoveEntryMemberUseCase(unitOfWork transaction.UnitOfWork, races race.Repository, entries participation.EntryRepository, runners RunnerRemover, teamID shared.TeamID) *RemoveEntryMemberUseCase {
	return &RemoveEntryMemberUseCase{unitOfWork: unitOfWork, races: races, entries: entries, runners: runners, teamID: teamID}
}

// Execute removes the member and takes them off the running orders, in one unit of work.
// Removing a member who is not in the entry does nothing. A race of another team, or an
// entry of another race, is shared.ErrNotFound.
func (useCase *RemoveEntryMemberUseCase) Execute(ctx context.Context, command EntryMemberCommand) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		entry, err := findRaceEntry(ctx, useCase.races, useCase.entries, useCase.teamID, command)
		if err != nil {
			return err
		}
		if !entry.Remove(command.memberID) {
			return nil
		}
		if err := useCase.entries.Save(ctx, entry); err != nil {
			return err
		}
		return useCase.runners.RemoveRunner(ctx, command.raceID, command.memberID)
	})
}
