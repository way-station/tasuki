package auth

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ExchangeLoginCodeUseCase exchanges the one-time code of LINE Login's callback for a
// session (POST /session). The code can be exchanged once.
type ExchangeLoginCodeUseCase struct {
	codes          identity.LoginCodeRepository
	sessions       identity.SessionRepository
	sessionFactory identity.SessionFactory
	dtoFactory     SessionDTOFactory
	now            func() time.Time
}

// NewExchangeLoginCodeUseCase returns an ExchangeLoginCodeUseCase.
func NewExchangeLoginCodeUseCase(codes identity.LoginCodeRepository, sessions identity.SessionRepository, sessionFactory identity.SessionFactory, dtoFactory SessionDTOFactory, now func() time.Time) *ExchangeLoginCodeUseCase {
	return &ExchangeLoginCodeUseCase{codes: codes, sessions: sessions, sessionFactory: sessionFactory, dtoFactory: dtoFactory, now: now}
}

// Execute returns the new session. An unknown, expired, or already exchanged code is
// ErrInvalidLoginCode.
func (useCase *ExchangeLoginCodeUseCase) Execute(ctx context.Context, code string) (SessionDTO, error) {
	loginCode, err := useCase.codes.Take(ctx, identity.HashToken(code))
	if errors.Is(err, shared.ErrNotFound) {
		return SessionDTO{}, ErrInvalidLoginCode
	}
	if err != nil {
		return SessionDTO{}, err
	}
	now := useCase.now()
	if !loginCode.Usable(now) {
		return SessionDTO{}, ErrInvalidLoginCode
	}
	session, token, err := useCase.sessionFactory.New(loginCode.UserID(), now)
	if err != nil {
		return SessionDTO{}, err
	}
	if err := useCase.sessions.Save(ctx, session); err != nil {
		return SessionDTO{}, err
	}
	return useCase.dtoFactory.Create(token, session.ExpiresAt()), nil
}
