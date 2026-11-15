package postgres

import (
	"context"
	"time"

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

var (
	_ identity.SessionRepository     = (*SessionRepository)(nil)
	_ identity.UserSessionRepository = (*SessionRepository)(nil)
)

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
		shared.UserID(row.UserID.String()),
		row.TokenHash,
		row.ExpiresAt.Time,
		timeOrNil(row.RevokedAt),
	), nil
}

// Save creates the session, or records that it was revoked. It runs in the transaction of a
// unit of work when there is one.
func (repository *SessionRepository) Save(ctx context.Context, session *identity.Session) error {
	sessionID, err := parseID(session.ID())
	if err != nil {
		return err
	}
	userID, err := parseID(session.UserID())
	if err != nil {
		return err
	}
	var revokedAt pgtype.Timestamptz
	if at, ok := session.RevokedAt(); ok {
		revokedAt = timestamptz(at)
	}
	return queriesFor(ctx, repository.queries).SaveSession(ctx, sqlcgen.SaveSessionParams{
		ID:        sessionID,
		UserID:    userID,
		TokenHash: session.TokenHash(),
		ExpiresAt: timestamptz(session.ExpiresAt()),
		RevokedAt: revokedAt,
	})
}

// ListActiveByUser returns the user's sessions that are active at now, oldest first. It runs
// in the transaction of a unit of work when there is one.
func (repository *SessionRepository) ListActiveByUser(ctx context.Context, userID shared.UserID, now time.Time) ([]*identity.Session, error) {
	user, err := parseID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := queriesFor(ctx, repository.queries).ListActiveSessionsByUser(ctx, sqlcgen.ListActiveSessionsByUserParams{UserID: user, Now: timestamptz(now)})
	if err != nil {
		return nil, err
	}
	sessions := make([]*identity.Session, 0, len(rows))
	for _, row := range rows {
		sessions = append(sessions, repository.factory.Reconstruct(shared.SessionID(row.ID.String()), shared.UserID(row.UserID.String()), row.TokenHash, row.ExpiresAt.Time, timeOrNil(row.RevokedAt)))
	}
	return sessions, nil
}
