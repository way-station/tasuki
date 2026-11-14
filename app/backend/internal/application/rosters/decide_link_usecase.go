package rosters

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// DecideLinkUseCase approves or rejects a link that waits for approval
// (POST /member-links/{link_id}/approve and /reject). Only organizers may call it; the
// middleware checks the role.
type DecideLinkUseCase struct {
	unitOfWork transaction.UnitOfWork
	members    roster.MemberRepository
	links      identity.MemberLinkRepository
	now        func() time.Time
	teamID     shared.TeamID
}

// NewDecideLinkUseCase returns a DecideLinkUseCase for the team.
func NewDecideLinkUseCase(unitOfWork transaction.UnitOfWork, members roster.MemberRepository, links identity.MemberLinkRepository, now func() time.Time, teamID shared.TeamID) *DecideLinkUseCase {
	return &DecideLinkUseCase{unitOfWork: unitOfWork, members: members, links: links, now: now, teamID: teamID}
}

// Approve approves the link. A link that does not wait for approval is a
// shared.StateError, and a link of a member of another team is shared.ErrNotFound.
func (useCase *DecideLinkUseCase) Approve(ctx context.Context, linkID shared.MemberLinkID) error {
	return useCase.decide(ctx, linkID, (*identity.MemberLink).Approve)
}

// Reject rejects the link. The organizer tells the person outside the app. A link that does
// not wait for approval is a shared.StateError, and a link of a member of another team is
// shared.ErrNotFound.
func (useCase *DecideLinkUseCase) Reject(ctx context.Context, linkID shared.MemberLinkID) error {
	return useCase.decide(ctx, linkID, (*identity.MemberLink).Reject)
}

func (useCase *DecideLinkUseCase) decide(ctx context.Context, linkID shared.MemberLinkID, decision func(link *identity.MemberLink, now time.Time) error) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		link, err := useCase.links.FindByID(ctx, linkID)
		if err != nil {
			return err
		}
		if _, err := findTeamMember(ctx, useCase.members, useCase.teamID, link.MemberID()); err != nil {
			return err
		}
		if err := decision(link, useCase.now()); err != nil {
			return err
		}
		return useCase.links.Save(ctx, link)
	})
}
