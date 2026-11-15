package participations

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ListInvitationsUseCase lists the roster members with their invitation to a race
// (GET /races/{race_id}/invitations). Only organizers may call it; the middleware checks
// the role.
type ListInvitationsUseCase struct {
	query      InvitationQueryService
	policy     participation.AttendancePolicy
	dtoFactory ParticipationDTOFactory
	teamID     shared.TeamID
}

// NewListInvitationsUseCase returns a ListInvitationsUseCase for the team.
func NewListInvitationsUseCase(query InvitationQueryService, policy participation.AttendancePolicy, dtoFactory ParticipationDTOFactory, teamID shared.TeamID) *ListInvitationsUseCase {
	return &ListInvitationsUseCase{query: query, policy: policy, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute returns every roster member who has not left. A race of another team is
// shared.ErrNotFound.
func (useCase *ListInvitationsUseCase) Execute(ctx context.Context, caller auth.MeDTO, raceID shared.RaceID) ([]InvitationDTO, error) {
	records, err := useCase.query.ListInvitations(ctx, useCase.teamID, raceID)
	if err != nil {
		return nil, err
	}
	canSeeAnswers := useCase.policy.CanSeeAllAnswers(caller.IsOrganizer())
	invitations := make([]InvitationDTO, 0, len(records))
	for _, record := range records {
		invitations = append(invitations, useCase.dtoFactory.CreateInvitation(record.MemberID, record.Name, record.Invited, record.Attendance, canSeeAnswers))
	}
	return invitations, nil
}
