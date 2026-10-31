package auth

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LoginUseCase starts a session for a verified identity: it creates the account on the
// first login, keeps its display name up to date, and creates the session. Only how the
// identity is verified differs between the dev-only login and LINE Login (ADR-0009).
type LoginUseCase struct {
	accounts       identity.AccountRepository
	sessions       identity.SessionRepository
	accountFactory identity.AccountFactory
	sessionFactory identity.SessionFactory
	dtoFactory     SessionDTOFactory
	now            func() time.Time
}

// NewLoginUseCase returns a LoginUseCase.
func NewLoginUseCase(accounts identity.AccountRepository, sessions identity.SessionRepository, accountFactory identity.AccountFactory, sessionFactory identity.SessionFactory, dtoFactory SessionDTOFactory, now func() time.Time) *LoginUseCase {
	return &LoginUseCase{accounts: accounts, sessions: sessions, accountFactory: accountFactory, sessionFactory: sessionFactory, dtoFactory: dtoFactory, now: now}
}

// Execute logs in as the command's LINE user and returns the new session.
func (useCase *LoginUseCase) Execute(ctx context.Context, command LoginCommand) (SessionDTO, error) {
	account, err := useCase.accounts.FindByLineUserID(ctx, command.lineUserID)
	switch {
	case errors.Is(err, shared.ErrNotFound):
		account, err = useCase.accountFactory.New(command.lineUserID, command.displayName)
		if err != nil {
			return SessionDTO{}, err
		}
	case err != nil:
		return SessionDTO{}, err
	default:
		account.UpdateDisplayName(command.displayName)
	}
	if err := useCase.accounts.Save(ctx, account); err != nil {
		return SessionDTO{}, err
	}

	session, token, err := useCase.sessionFactory.New(account.ID(), useCase.now())
	if err != nil {
		return SessionDTO{}, err
	}
	if err := useCase.sessions.Save(ctx, session); err != nil {
		return SessionDTO{}, err
	}
	return useCase.dtoFactory.Create(token, session.ExpiresAt()), nil
}
