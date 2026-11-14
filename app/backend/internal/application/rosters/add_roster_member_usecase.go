package rosters

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// AddRosterMemberUseCase adds a member to the roster (POST /roster-members). Only organizers
// may call it; the middleware checks the role.
type AddRosterMemberUseCase struct {
	members    roster.MemberRepository
	factory    roster.MemberFactory
	dtoFactory RosterDTOFactory
	teamID     shared.TeamID
}

// NewAddRosterMemberUseCase returns an AddRosterMemberUseCase for the team.
func NewAddRosterMemberUseCase(members roster.MemberRepository, factory roster.MemberFactory, dtoFactory RosterDTOFactory, teamID shared.TeamID) *AddRosterMemberUseCase {
	return &AddRosterMemberUseCase{members: members, factory: factory, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute adds the member and returns them. A blank name is a shared.ValidationError.
func (useCase *AddRosterMemberUseCase) Execute(ctx context.Context, command AddRosterMemberCommand) (RosterMemberDTO, error) {
	member, err := useCase.factory.New(useCase.teamID, command.name)
	if err != nil {
		return RosterMemberDTO{}, err
	}
	if err := useCase.members.Save(ctx, member); err != nil {
		return RosterMemberDTO{}, err
	}
	return useCase.dtoFactory.CreateRosterMember(member.ID(), member.Name(), member.Left(), roster.LinkNone, nil), nil
}
