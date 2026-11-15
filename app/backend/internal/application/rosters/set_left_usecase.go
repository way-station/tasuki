package rosters

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// SetLeftUseCase marks a roster member as left, or undoes it
// (PATCH /roster-members/{roster_member_id}). Only organizers may call it; the middleware
// checks the role.
type SetLeftUseCase struct {
	unitOfWork   transaction.UnitOfWork
	members      roster.MemberRepository
	query        RosterQueryService
	invitations  InvitationCanceller
	links        identity.MemberLinkRepository
	userSessions identity.UserSessionRepository
	sessions     identity.SessionRepository
	leave        roster.LeaveService
	now          func() time.Time
	location     *time.Location
	teamID       shared.TeamID
}

// NewSetLeftUseCase returns a SetLeftUseCase for the team. Today is the date at now in
// location (Japan time, docs/design/api.md).
func NewSetLeftUseCase(unitOfWork transaction.UnitOfWork, members roster.MemberRepository, query RosterQueryService, invitations InvitationCanceller, links identity.MemberLinkRepository, userSessions identity.UserSessionRepository, sessions identity.SessionRepository, leave roster.LeaveService, now func() time.Time, location *time.Location, teamID shared.TeamID) *SetLeftUseCase {
	return &SetLeftUseCase{unitOfWork: unitOfWork, members: members, query: query, invitations: invitations, links: links, userSessions: userSessions, sessions: sessions, leave: leave, now: now, location: location, teamID: teamID}
}

// Execute marks the member as left, or undoes it, in one unit of work. Leaving cancels the
// invitations to races dated today or later and revokes every active session of the user
// linked to the member. Marking a left member again, or undoing it for a member who has
// not left, changes nothing. A member of another team is shared.ErrNotFound. The response
// is read with ListRosterMembersUseCase.Find.
func (useCase *SetLeftUseCase) Execute(ctx context.Context, command SetLeftCommand) error {
	return useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		member, err := findTeamMember(ctx, useCase.members, useCase.teamID, command.memberID)
		if err != nil {
			return err
		}
		if member.Left() == command.left {
			return nil
		}
		now := useCase.now()
		if !command.left {
			member.Rejoin()
			return useCase.members.Save(ctx, member)
		}
		member.Leave(now)
		if err := useCase.members.Save(ctx, member); err != nil {
			return err
		}
		if err := useCase.cancelInvitations(ctx, member.ID(), shared.DateIn(now, useCase.location)); err != nil {
			return err
		}
		return useCase.revokeSessions(ctx, member.ID(), now)
	})
}

// cancelInvitations cancels the member's invitations to races dated today or later.
func (useCase *SetLeftUseCase) cancelInvitations(ctx context.Context, memberID shared.MemberID, today shared.Date) error {
	invited, err := useCase.query.ListInvitedRaces(ctx, memberID)
	if err != nil {
		return err
	}
	for _, raceID := range useCase.leave.InvitationsToCancel(today, invited) {
		if err := useCase.invitations.CancelInvitation(ctx, raceID, memberID); err != nil {
			return err
		}
	}
	return nil
}

// revokeSessions revokes the active sessions of the user linked to the member, if any.
func (useCase *SetLeftUseCase) revokeSessions(ctx context.Context, memberID shared.MemberID, now time.Time) error {
	links, err := useCase.links.ListByMember(ctx, memberID)
	if err != nil {
		return err
	}
	for _, link := range links {
		if !link.Active() {
			continue
		}
		sessions, err := useCase.userSessions.ListActiveByUser(ctx, link.UserID(), now)
		if err != nil {
			return err
		}
		for _, session := range sessions {
			session.Revoke(now)
			if err := useCase.sessions.Save(ctx, session); err != nil {
				return err
			}
		}
	}
	return nil
}
