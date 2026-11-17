package auth

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// StartLineLoginUseCase starts LINE Login (POST /auth/line/authorize). It keeps the state,
// the nonce, the PKCE verifier, and the invite link for the callback (ADR-0007).
type StartLineLoginUseCase struct {
	inviteLinks    identity.InviteLinkRepository
	attempts       identity.LoginAttemptRepository
	attemptFactory identity.LoginAttemptFactory
	lineLogin      LineLogin
	now            func() time.Time
}

// NewStartLineLoginUseCase returns a StartLineLoginUseCase.
func NewStartLineLoginUseCase(inviteLinks identity.InviteLinkRepository, attempts identity.LoginAttemptRepository, attemptFactory identity.LoginAttemptFactory, lineLogin LineLogin, now func() time.Time) *StartLineLoginUseCase {
	return &StartLineLoginUseCase{inviteLinks: inviteLinks, attempts: attempts, attemptFactory: attemptFactory, lineLogin: lineLogin, now: now}
}

// Execute returns the URL of LINE's login page. An invite token that is unknown, expired,
// used, or replaced is identity.ErrInviteLinkUnusable.
func (useCase *StartLineLoginUseCase) Execute(ctx context.Context, command StartLineLoginCommand) (string, error) {
	var invite *identity.InviteLink
	if token, ok := command.InviteToken(); ok {
		link, err := useCase.inviteLinks.FindByTokenHash(ctx, identity.HashToken(token))
		if errors.Is(err, shared.ErrNotFound) {
			return "", identity.ErrInviteLinkUnusable
		}
		if err != nil {
			return "", err
		}
		invite = link
	}
	attempt, state, err := useCase.attemptFactory.New(invite, useCase.now())
	if err != nil {
		return "", err
	}
	if err := useCase.attempts.Save(ctx, attempt); err != nil {
		return "", err
	}
	return useCase.lineLogin.AuthorizationURL(state, attempt.Nonce(), attempt.CodeVerifier()), nil
}
