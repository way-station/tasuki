package auth

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// AuthenticateUseCase identifies the account of a session token (ADR-0014).
type AuthenticateUseCase struct {
	sessions SessionQueryService
	now      func() time.Time
}

// NewAuthenticateUseCase returns an AuthenticateUseCase that reads through sessions.
func NewAuthenticateUseCase(sessions SessionQueryService, now func() time.Time) *AuthenticateUseCase {
	return &AuthenticateUseCase{sessions: sessions, now: now}
}

// Execute returns the account of token, or ErrUnauthenticated when the token is empty,
// unknown, expired, or revoked.
func (useCase *AuthenticateUseCase) Execute(ctx context.Context, token string) (shared.AccountID, error) {
	if token == "" {
		return "", ErrUnauthenticated
	}
	accountID, err := useCase.sessions.FindActiveAccountID(ctx, identity.HashToken(token), useCase.now())
	if errors.Is(err, shared.ErrNotFound) {
		return "", ErrUnauthenticated
	}
	return accountID, err
}
