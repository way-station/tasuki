package auth

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LogoutUseCase revokes the caller's session (DELETE /session).
type LogoutUseCase struct {
	sessions identity.SessionRepository
	now      func() time.Time
}

// NewLogoutUseCase returns a LogoutUseCase.
func NewLogoutUseCase(sessions identity.SessionRepository, now func() time.Time) *LogoutUseCase {
	return &LogoutUseCase{sessions: sessions, now: now}
}

// Execute revokes the session of token. A session that is already gone is not an error.
func (useCase *LogoutUseCase) Execute(ctx context.Context, token string) error {
	session, err := useCase.sessions.FindByTokenHash(ctx, identity.HashToken(token))
	if errors.Is(err, shared.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	session.Revoke(useCase.now())
	return useCase.sessions.Save(ctx, session)
}
