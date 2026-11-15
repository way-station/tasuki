package rosters

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// UnlinkUseCase removes the approved link of a roster member
// (DELETE /roster-members/{roster_member_id}/link). Only organizers may call it; the
// middleware checks the role.
type UnlinkUseCase struct {
	unitOfWork transaction.UnitOfWork
	members    roster.MemberRepository
	links      identity.MemberLinkRepository
	now        func() time.Time
	teamID     shared.TeamID
}

// NewUnlinkUseCase returns an UnlinkUseCase for the team.
func NewUnlinkUseCase(unitOfWork transaction.UnitOfWork, members roster.MemberRepository, links identity.MemberLinkRepository, now func() time.Time, teamID shared.TeamID) *UnlinkUseCase {
	return &UnlinkUseCase{unitOfWork: unitOfWork, members: members, links: links, now: now, teamID: teamID}
}

// Execute unlinks the member's approved link. The user's sessions stay valid. A member
// without an approved link is a shared.StateError, and a member of another team is
// shared.ErrNotFound.
func (useCase *UnlinkUseCase) Execute(ctx context.Context, memberID shared.MemberID) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		if _, err := findTeamMember(ctx, useCase.members, useCase.teamID, memberID); err != nil {
			return err
		}
		links, err := useCase.links.ListByMember(ctx, memberID)
		if err != nil {
			return err
		}
		for _, link := range links {
			if link.Active() {
				if err := link.Unlink(useCase.now()); err != nil {
					return err
				}
				return useCase.links.Save(ctx, link)
			}
		}
		return shared.NewStateError("the member has no approved link")
	})
}
