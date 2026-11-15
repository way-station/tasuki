package auth

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// CompleteLineLoginUseCase finishes LINE Login on LINE's callback (GET /auth/line/callback):
// it verifies the ID token (ADR-0007), creates or updates the user, links the user to the
// invite link's roster member when the login started from one, and returns a one-time code
// for POST /session. The writes run in one transaction.
type CompleteLineLoginUseCase struct {
	unitOfWork    transaction.UnitOfWork
	attempts      identity.LoginAttemptRepository
	lineLogin     LineLogin
	users         identity.UserRepository
	userFactory   identity.UserFactory
	memberLinks   identity.MemberLinkRepository
	linkFactory   identity.MemberLinkFactory
	inviteQuery   InviteLinkQueryService
	inviteLinks   identity.InviteLinkRepository
	inviteFactory identity.InviteLinkFactory
	members       roster.MemberRepository
	codes         identity.LoginCodeRepository
	codeFactory   identity.LoginCodeFactory
	now           func() time.Time
}

// NewCompleteLineLoginUseCase returns a CompleteLineLoginUseCase.
func NewCompleteLineLoginUseCase(unitOfWork transaction.UnitOfWork, attempts identity.LoginAttemptRepository, lineLogin LineLogin, users identity.UserRepository, userFactory identity.UserFactory, memberLinks identity.MemberLinkRepository, linkFactory identity.MemberLinkFactory, inviteQuery InviteLinkQueryService, inviteLinks identity.InviteLinkRepository, inviteFactory identity.InviteLinkFactory, members roster.MemberRepository, codes identity.LoginCodeRepository, codeFactory identity.LoginCodeFactory, now func() time.Time) *CompleteLineLoginUseCase {
	return &CompleteLineLoginUseCase{unitOfWork: unitOfWork, attempts: attempts, lineLogin: lineLogin, users: users, userFactory: userFactory, memberLinks: memberLinks, linkFactory: linkFactory, inviteQuery: inviteQuery, inviteLinks: inviteLinks, inviteFactory: inviteFactory, members: members, codes: codes, codeFactory: codeFactory, now: now}
}

// Execute returns the one-time code. It fails with
//   - ErrLoginFailed when LINE set an error or no code, the state is unknown or already
//     used, the login took longer than identity.LoginAttemptLifetime, or the ID token is
//     not valid;
//   - ErrMemberLeft when the roster member of the user's approved link, or of the invite
//     link, has left;
//   - identity.ErrInviteLinkUnusable when the invite link became unusable after the login
//     started.
//
// A user who already has an approved or pending link gets no new link, and the invite link
// stays usable.
func (useCase *CompleteLineLoginUseCase) Execute(ctx context.Context, command CompleteLineLoginCommand) (string, error) {
	attempt, err := useCase.attempts.Take(ctx, identity.HashToken(command.state))
	if errors.Is(err, shared.ErrNotFound) {
		return "", ErrLoginFailed
	}
	if err != nil {
		return "", err
	}
	now := useCase.now()
	if command.lineError != "" || command.code == "" || !attempt.Usable(now) {
		return "", ErrLoginFailed
	}
	lineIdentity, err := useCase.lineLogin.Authenticate(ctx, command.code, attempt.CodeVerifier(), attempt.Nonce())
	if err != nil {
		return "", errors.Join(ErrLoginFailed, err)
	}

	var code string
	err = useCase.unitOfWork.Run(ctx, func(ctx context.Context) error {
		user, err := useCase.saveUser(ctx, lineIdentity)
		if err != nil {
			return err
		}
		links, err := useCase.memberLinks.ListByUser(ctx, user.ID())
		if err != nil {
			return err
		}
		if err := useCase.refuseLeftMember(ctx, links); err != nil {
			return err
		}
		if inviteLinkID, ok := attempt.InviteLinkID(); ok && !hasOpenLink(links) {
			if err := useCase.linkToInvite(ctx, user.ID(), inviteLinkID, now); err != nil {
				return err
			}
		}
		loginCode, plain, err := useCase.codeFactory.New(user.ID(), now)
		if err != nil {
			return err
		}
		if err := useCase.codes.Save(ctx, loginCode); err != nil {
			return err
		}
		code = plain
		return nil
	})
	if err != nil {
		return "", err
	}
	return code, nil
}

// saveUser creates the user on the first login and keeps its display name up to date.
func (useCase *CompleteLineLoginUseCase) saveUser(ctx context.Context, lineIdentity LineIdentity) (*identity.User, error) {
	user, err := useCase.users.FindByLineUserID(ctx, lineIdentity.UserID)
	switch {
	case errors.Is(err, shared.ErrNotFound):
		user, err = useCase.userFactory.New(lineIdentity.UserID, lineIdentity.DisplayName)
		if err != nil {
			return nil, err
		}
	case err != nil:
		return nil, err
	default:
		user.UpdateDisplayName(lineIdentity.DisplayName)
	}
	if err := useCase.users.Save(ctx, user); err != nil {
		return nil, err
	}
	return user, nil
}

// refuseLeftMember returns ErrMemberLeft when the roster member of an approved link has left.
func (useCase *CompleteLineLoginUseCase) refuseLeftMember(ctx context.Context, links []*identity.MemberLink) error {
	for _, link := range links {
		if !link.Active() {
			continue
		}
		if err := useCase.refuseIfLeft(ctx, link.MemberID()); err != nil {
			return err
		}
	}
	return nil
}

// linkToInvite creates a pending link between the user and the invite link's roster member,
// and makes the invite link unusable.
func (useCase *CompleteLineLoginUseCase) linkToInvite(ctx context.Context, userID shared.UserID, inviteLinkID shared.InviteLinkID, now time.Time) error {
	record, err := useCase.inviteQuery.FindByID(ctx, inviteLinkID)
	if errors.Is(err, shared.ErrNotFound) {
		return identity.ErrInviteLinkUnusable
	}
	if err != nil {
		return err
	}
	invite := useCase.inviteFactory.Reconstruct(record.ID, record.MemberID, record.TokenHash, record.ExpiresAt, record.UsedAt, record.RevokedAt)
	if err := useCase.refuseIfLeft(ctx, invite.MemberID()); err != nil {
		return err
	}
	link, err := useCase.linkFactory.New(userID, invite, now)
	if err != nil {
		return err
	}
	if err := useCase.memberLinks.Save(ctx, link); err != nil {
		return err
	}
	return useCase.inviteLinks.Save(ctx, invite)
}

func (useCase *CompleteLineLoginUseCase) refuseIfLeft(ctx context.Context, memberID shared.MemberID) error {
	member, err := useCase.members.FindByID(ctx, memberID)
	if err != nil {
		return err
	}
	if member.Left() {
		return ErrMemberLeft
	}
	return nil
}

// hasOpenLink reports whether one of the links is approved or waits for approval.
func hasOpenLink(links []*identity.MemberLink) bool {
	for _, link := range links {
		if link.Active() || link.Pending() {
			return true
		}
	}
	return false
}
