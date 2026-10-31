package auth

import (
	"context"
	"errors"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// LoginUseCase starts a session for a verified identity: it creates the user on the
// first login, keeps its display name up to date, and creates the session. Only how the
// identity is verified differs between the dev-only login and LINE Login (ADR-0009).
type LoginUseCase struct {
	users          identity.UserRepository
	sessions       identity.SessionRepository
	userFactory    identity.UserFactory
	sessionFactory identity.SessionFactory
	dtoFactory     SessionDTOFactory
	now            func() time.Time
}

// NewLoginUseCase returns a LoginUseCase.
func NewLoginUseCase(users identity.UserRepository, sessions identity.SessionRepository, userFactory identity.UserFactory, sessionFactory identity.SessionFactory, dtoFactory SessionDTOFactory, now func() time.Time) *LoginUseCase {
	return &LoginUseCase{users: users, sessions: sessions, userFactory: userFactory, sessionFactory: sessionFactory, dtoFactory: dtoFactory, now: now}
}

// Execute logs in as the command's LINE user and returns the new session.
func (useCase *LoginUseCase) Execute(ctx context.Context, command LoginCommand) (SessionDTO, error) {
	user, err := useCase.users.FindByLineUserID(ctx, command.lineUserID)
	switch {
	case errors.Is(err, shared.ErrNotFound):
		user, err = useCase.userFactory.New(command.lineUserID, command.displayName)
		if err != nil {
			return SessionDTO{}, err
		}
	case err != nil:
		return SessionDTO{}, err
	default:
		user.UpdateDisplayName(command.displayName)
	}
	if err := useCase.users.Save(ctx, user); err != nil {
		return SessionDTO{}, err
	}

	session, token, err := useCase.sessionFactory.New(user.ID(), useCase.now())
	if err != nil {
		return SessionDTO{}, err
	}
	if err := useCase.sessions.Save(ctx, session); err != nil {
		return SessionDTO{}, err
	}
	return useCase.dtoFactory.Create(token, session.ExpiresAt()), nil
}
