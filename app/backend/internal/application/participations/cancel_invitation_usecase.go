package participations

import (
	"context"
	"errors"

	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// CancelInvitationUseCase cancels the invitation of a roster member to a race
// (DELETE /races/{race_id}/invitations/{roster_member_id}). Only organizers may call it;
// the middleware checks the role.
type CancelInvitationUseCase struct {
	unitOfWork  transaction.UnitOfWork
	races       race.Repository
	invitations participation.InvitationRepository
	entries     participation.EntryRepository
	runners     RunnerRemover
	teamID      shared.TeamID
}

// NewCancelInvitationUseCase returns a CancelInvitationUseCase for the team.
func NewCancelInvitationUseCase(unitOfWork transaction.UnitOfWork, races race.Repository, invitations participation.InvitationRepository, entries participation.EntryRepository, runners RunnerRemover, teamID shared.TeamID) *CancelInvitationUseCase {
	return &CancelInvitationUseCase{unitOfWork: unitOfWork, races: races, invitations: invitations, entries: entries, runners: runners, teamID: teamID}
}

// Execute deletes the invitation with the answer, takes the member out of the entry, and
// takes them off the running orders, in one unit of work. Cancelling the invitation of a
// member who is not invited does nothing. A race of another team is shared.ErrNotFound.
func (useCase *CancelInvitationUseCase) Execute(ctx context.Context, command InvitationCommand) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		if err := requireTeamRace(ctx, useCase.races, useCase.teamID, command.raceID); err != nil {
			return err
		}
		_, err := useCase.invitations.Find(ctx, command.raceID, command.memberID)
		if errors.Is(err, shared.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := leaveEntries(ctx, useCase.entries, useCase.runners, command.raceID, command.memberID); err != nil {
			return err
		}
		return useCase.invitations.Delete(ctx, command.raceID, command.memberID)
	})
}
