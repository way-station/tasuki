package postgres

import (
	"context"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// SessionQueryService implements auth.SessionQueryService with PostgreSQL.
type SessionQueryService struct {
	queries *sqlcgen.Queries
}

var _ auth.SessionQueryService = (*SessionQueryService)(nil)

// NewSessionQueryService returns a SessionQueryService that uses client.
func NewSessionQueryService(client *Client) *SessionQueryService {
	return &SessionQueryService{queries: sqlcgen.New(client.Pool())}
}

// FindActiveAccountID returns the account of the active session whose token has hash.
func (queryService *SessionQueryService) FindActiveAccountID(ctx context.Context, hash []byte, now time.Time) (shared.AccountID, error) {
	accountID, err := queryService.queries.GetActiveSessionAccountID(ctx, sqlcgen.GetActiveSessionAccountIDParams{
		TokenHash: hash,
		Now:       timestamptz(now),
	})
	if err != nil {
		return "", notFound(err, "session")
	}
	return shared.AccountID(accountID.String()), nil
}
