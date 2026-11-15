package rosters

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// CreateInviteLinkUseCase creates an invite link for a roster member
// (POST /roster-members/{roster_member_id}/invite-links). Only organizers may call it; the
// middleware checks the role.
type CreateInviteLinkUseCase struct {
	unitOfWork  transaction.UnitOfWork
	members     roster.MemberRepository
	links       identity.MemberLinkRepository
	inviteLinks identity.InviteLinkRepository
	factory     identity.InviteLinkFactory
	linkStatus  identity.LinkStatusService
	dtoFactory  RosterDTOFactory
	now         func() time.Time
	teamID      shared.TeamID
}

// NewCreateInviteLinkUseCase returns a CreateInviteLinkUseCase for the team.
func NewCreateInviteLinkUseCase(unitOfWork transaction.UnitOfWork, members roster.MemberRepository, links identity.MemberLinkRepository, inviteLinks identity.InviteLinkRepository, factory identity.InviteLinkFactory, linkStatus identity.LinkStatusService, dtoFactory RosterDTOFactory, now func() time.Time, teamID shared.TeamID) *CreateInviteLinkUseCase {
	return &CreateInviteLinkUseCase{unitOfWork: unitOfWork, members: members, links: links, inviteLinks: inviteLinks, factory: factory, linkStatus: linkStatus, dtoFactory: dtoFactory, now: now, teamID: teamID}
}

// Execute replaces the member's previous invite link with a new one and returns its token.
// A member who has left, or has an approved or pending link, is a shared.StateError, and a
// member of another team is shared.ErrNotFound.
func (useCase *CreateInviteLinkUseCase) Execute(ctx context.Context, memberID shared.MemberID) (InviteLinkDTO, error) {
	var created InviteLinkDTO
	err := useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		member, err := findTeamMember(ctx, useCase.members, useCase.teamID, memberID)
		if err != nil {
			return err
		}
		links, err := useCase.links.ListByMember(ctx, memberID)
		if err != nil {
			return err
		}
		if err := member.CanIssueInviteLink(useCase.linkStatus.RosterLinkStatus(links)); err != nil {
			return err
		}
		now := useCase.now()
		if err := useCase.revokePrevious(ctx, memberID, now); err != nil {
			return err
		}
		invite, token, err := useCase.factory.New(memberID, now)
		if err != nil {
			return err
		}
		if err := useCase.inviteLinks.Save(ctx, invite); err != nil {
			return err
		}
		created = useCase.dtoFactory.CreateInviteLink(token, invite.ExpiresAt())
		return nil
	})
	return created, err
}

// revokePrevious makes the member's previous invite link unusable, if there is one.
func (useCase *CreateInviteLinkUseCase) revokePrevious(ctx context.Context, memberID shared.MemberID, now time.Time) error {
	previous, err := useCase.inviteLinks.FindOpenByMember(ctx, memberID)
	if errors.Is(err, shared.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	previous.Revoke(now)
	return useCase.inviteLinks.Save(ctx, previous)
}
