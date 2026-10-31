package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// SessionRepository implements identity.SessionRepository with PostgreSQL.
type SessionRepository struct {
	queries *sqlcgen.Queries
	factory identity.SessionFactory
}

var _ identity.SessionRepository = (*SessionRepository)(nil)

// NewSessionRepository returns a SessionRepository that uses client.
func NewSessionRepository(client *Client, factory identity.SessionFactory) *SessionRepository {
	return &SessionRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// FindByTokenHash returns the session whose token has hash, whatever its state.
func (repository *SessionRepository) FindByTokenHash(ctx context.Context, hash []byte) (*identity.Session, error) {
	row, err := repository.queries.GetSessionByTokenHash(ctx, hash)
	if err != nil {
		return nil, notFound(err, "session")
	}
	return repository.factory.Reconstruct(
		shared.SessionID(row.ID.String()),
		shared.AccountID(row.AccountID.String()),
		row.TokenHash,
		row.ExpiresAt.Time,
		timeOrNil(row.RevokedAt),
	), nil
}

// Save creates the session, or records that it was revoked.
func (repository *SessionRepository) Save(ctx context.Context, session *identity.Session) error {
	sessionID, err := parseID(session.ID())
	if err != nil {
		return err
	}
	accountID, err := parseID(session.AccountID())
	if err != nil {
		return err
	}
	var revokedAt pgtype.Timestamptz
	if at, ok := session.RevokedAt(); ok {
		revokedAt = timestamptz(at)
	}
	return repository.queries.SaveSession(ctx, sqlcgen.SaveSessionParams{
		ID:        sessionID,
		AccountID: accountID,
		TokenHash: session.TokenHash(),
		ExpiresAt: timestamptz(session.ExpiresAt()),
		RevokedAt: revokedAt,
	})
}
