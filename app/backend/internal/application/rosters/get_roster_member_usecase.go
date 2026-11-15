package rosters

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// GetRosterMemberUseCase returns a roster member with their average time and their laps by
// race (GET /roster-members/{roster_member_id}). Only organizers may call it; the
// middleware checks the role.
type GetRosterMemberUseCase struct {
	query      RosterQueryService
	dtoFactory RosterDTOFactory
	teamID     shared.TeamID
}

// NewGetRosterMemberUseCase returns a GetRosterMemberUseCase for the team.
func NewGetRosterMemberUseCase(query RosterQueryService, dtoFactory RosterDTOFactory, teamID shared.TeamID) *GetRosterMemberUseCase {
	return &GetRosterMemberUseCase{query: query, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute returns the member. A member of another team is shared.ErrNotFound.
func (useCase *GetRosterMemberUseCase) Execute(ctx context.Context, memberID shared.MemberID) (RosterMemberDetailDTO, error) {
	record, err := useCase.query.GetRosterMember(ctx, useCase.teamID, memberID)
	if err != nil {
		return RosterMemberDetailDTO{}, err
	}
	return useCase.dtoFactory.CreateRosterMemberDetail(record), nil
}
