package participations

import (
	"context"
	"errors"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// InviteUseCase invites a roster member to a race (PUT /races/{race_id}/invitations/
// {roster_member_id}). Only organizers may call it; the middleware checks the role.
type InviteUseCase struct {
	races       race.Repository
	members     roster.MemberRepository
	invitations participation.InvitationRepository
	factory     participation.InvitationFactory
	teamID      shared.TeamID
}

// NewInviteUseCase returns an InviteUseCase for the team.
func NewInviteUseCase(races race.Repository, members roster.MemberRepository, invitations participation.InvitationRepository, factory participation.InvitationFactory, teamID shared.TeamID) *InviteUseCase {
	return &InviteUseCase{races: races, members: members, invitations: invitations, factory: factory, teamID: teamID}
}

// Execute invites the member. Inviting an invited member does nothing, and a member who
// has left is a shared.ValidationError. A race or a member of another team is
// shared.ErrNotFound.
func (useCase *InviteUseCase) Execute(ctx context.Context, command InvitationCommand) error {
	if err := requireTeamRace(ctx, useCase.races, useCase.teamID, command.raceID); err != nil {
		return err
	}
	member, err := useCase.members.FindByID(ctx, command.memberID)
	if err != nil {
		return err
	}
	if member.TeamID() != useCase.teamID {
		return fmt.Errorf("roster member %s: %w", command.memberID, shared.ErrNotFound)
	}
	_, err = useCase.invitations.Find(ctx, command.raceID, command.memberID)
	if err == nil {
		return nil
	}
	if !errors.Is(err, shared.ErrNotFound) {
		return err
	}
	invitation, err := useCase.factory.New(command.raceID, command.memberID, member.Left())
	if err != nil {
		return err
	}
	return useCase.invitations.Save(ctx, invitation)
}
