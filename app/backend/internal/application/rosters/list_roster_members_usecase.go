package rosters

import (
	"context"
	"fmt"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ListRosterMembersUseCase lists the roster (GET /roster-members). Only organizers may call
// it; the middleware checks the role.
type ListRosterMembersUseCase struct {
	query             RosterQueryService
	linkFactory       identity.MemberLinkFactory
	inviteLinkFactory identity.InviteLinkFactory
	linkStatus        identity.LinkStatusService
	dtoFactory        RosterDTOFactory
	now               func() time.Time
	teamID            shared.TeamID
}

// NewListRosterMembersUseCase returns a ListRosterMembersUseCase for the team.
func NewListRosterMembersUseCase(query RosterQueryService, linkFactory identity.MemberLinkFactory, inviteLinkFactory identity.InviteLinkFactory, linkStatus identity.LinkStatusService, dtoFactory RosterDTOFactory, now func() time.Time, teamID shared.TeamID) *ListRosterMembersUseCase {
	return &ListRosterMembersUseCase{query: query, linkFactory: linkFactory, inviteLinkFactory: inviteLinkFactory, linkStatus: linkStatus, dtoFactory: dtoFactory, now: now, teamID: teamID}
}

// Execute returns the roster members with the state of their links and the expiry of their
// invite link that can still be used.
func (useCase *ListRosterMembersUseCase) Execute(ctx context.Context) ([]RosterMemberDTO, error) {
	records, err := useCase.query.ListRosterMembers(ctx, useCase.teamID)
	if err != nil {
		return nil, err
	}
	now := useCase.now()
	members := make([]RosterMemberDTO, 0, len(records))
	for _, record := range records {
		members = append(members, useCase.dtoOf(record, now))
	}
	return members, nil
}

// Find returns one roster member as the list shows them. The endpoints that change a member
// return the member through it. A member of another team is shared.ErrNotFound.
func (useCase *ListRosterMembersUseCase) Find(ctx context.Context, memberID shared.MemberID) (RosterMemberDTO, error) {
	members, err := useCase.query.ListRosterMembers(ctx, useCase.teamID)
	if err != nil {
		return RosterMemberDTO{}, err
	}
	for _, record := range members {
		if record.MemberID == memberID {
			return useCase.dtoOf(record, useCase.now()), nil
		}
	}
	return RosterMemberDTO{}, fmt.Errorf("roster member %s: %w", memberID, shared.ErrNotFound)
}

func (useCase *ListRosterMembersUseCase) dtoOf(record RosterMemberRecord, now time.Time) RosterMemberDTO {
	links := make([]*identity.MemberLink, 0, len(record.Links))
	for _, link := range record.Links {
		links = append(links, useCase.linkFactory.Reconstruct(link.ID, link.UserID, record.MemberID, link.InviteLinkID, link.CreatedAt, link.ApprovedAt, link.RejectedAt, link.UnlinkedAt))
	}
	var expiresAt *time.Time
	if record.InviteLink != nil {
		invite := useCase.inviteLinkFactory.Reconstruct(record.InviteLink.ID, record.MemberID, nil, record.InviteLink.ExpiresAt, nil, nil)
		if invite.Usable(now) {
			at := invite.ExpiresAt()
			expiresAt = &at
		}
	}
	return useCase.dtoFactory.CreateRosterMember(record.MemberID, record.Name, record.Left, useCase.linkStatus.RosterLinkStatus(links), expiresAt)
}
