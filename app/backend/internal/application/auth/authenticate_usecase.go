package auth

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// AuthenticateUseCase identifies the user of a session token (ADR-0014).
type AuthenticateUseCase struct {
	sessions SessionQueryService
	now      func() time.Time
}

// NewAuthenticateUseCase returns an AuthenticateUseCase that reads through sessions.
func NewAuthenticateUseCase(sessions SessionQueryService, now func() time.Time) *AuthenticateUseCase {
	return &AuthenticateUseCase{sessions: sessions, now: now}
}

// Execute returns the user of token, or ErrUnauthenticated when the token is empty,
// unknown, expired, or revoked.
func (useCase *AuthenticateUseCase) Execute(ctx context.Context, token string) (shared.UserID, error) {
	if token == "" {
		return "", ErrUnauthenticated
	}
	userID, err := useCase.sessions.FindActiveUserID(ctx, identity.HashToken(token), useCase.now())
	if errors.Is(err, shared.ErrNotFound) {
		return "", ErrUnauthenticated
	}
	return userID, err
}
