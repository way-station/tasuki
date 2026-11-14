package postgres

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// LoginCodeRepository implements identity.LoginCodeRepository with PostgreSQL. It runs in
// the transaction of a unit of work when there is one.
type LoginCodeRepository struct {
	queries *sqlcgen.Queries
	factory identity.LoginCodeFactory
}

var _ identity.LoginCodeRepository = (*LoginCodeRepository)(nil)

// NewLoginCodeRepository returns a LoginCodeRepository that uses client.
func NewLoginCodeRepository(client *Client, factory identity.LoginCodeFactory) *LoginCodeRepository {
	return &LoginCodeRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// Save creates the code. A code never changes; it is only taken.
func (repository *LoginCodeRepository) Save(ctx context.Context, code *identity.LoginCode) error {
	codeID, err := parseID(code.ID())
	if err != nil {
		return err
	}
	userID, err := parseID(code.UserID())
	if err != nil {
		return err
	}
	return queriesFor(ctx, repository.queries).InsertLoginCode(ctx, sqlcgen.InsertLoginCodeParams{
		ID:        codeID,
		UserID:    userID,
		CodeHash:  code.CodeHash(),
		ExpiresAt: timestamptz(code.ExpiresAt()),
	})
}

// Take deletes and returns the code whose hash is codeHash.
func (repository *LoginCodeRepository) Take(ctx context.Context, codeHash []byte) (*identity.LoginCode, error) {
	row, err := queriesFor(ctx, repository.queries).TakeLoginCode(ctx, codeHash)
	if err != nil {
		return nil, notFound(err, "login code")
	}
	return repository.factory.Reconstruct(
		shared.LoginCodeID(row.ID.String()),
		shared.UserID(row.UserID.String()),
		row.CodeHash,
		row.ExpiresAt.Time,
	), nil
}
