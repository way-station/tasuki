package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// LoginAttemptRepository implements identity.LoginAttemptRepository with PostgreSQL.
type LoginAttemptRepository struct {
	queries *sqlcgen.Queries
	factory identity.LoginAttemptFactory
}

var _ identity.LoginAttemptRepository = (*LoginAttemptRepository)(nil)

// NewLoginAttemptRepository returns a LoginAttemptRepository that uses client.
func NewLoginAttemptRepository(client *Client, factory identity.LoginAttemptFactory) *LoginAttemptRepository {
	return &LoginAttemptRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// Save creates the attempt. An attempt never changes; it is only taken.
func (repository *LoginAttemptRepository) Save(ctx context.Context, attempt *identity.LoginAttempt) error {
	attemptID, err := parseID(attempt.ID())
	if err != nil {
		return err
	}
	var inviteLinkID *uuid.UUID
	if id, ok := attempt.InviteLinkID(); ok {
		parsed, err := parseID(id)
		if err != nil {
			return err
		}
		inviteLinkID = &parsed
	}
	return repository.queries.InsertLoginAttempt(ctx, sqlcgen.InsertLoginAttemptParams{
		ID:           attemptID,
		StateHash:    attempt.StateHash(),
		Nonce:        attempt.Nonce(),
		CodeVerifier: attempt.CodeVerifier(),
		InviteLinkID: inviteLinkID,
		ExpiresAt:    timestamptz(attempt.ExpiresAt()),
	})
}

// Take deletes and returns the attempt whose state has hash.
func (repository *LoginAttemptRepository) Take(ctx context.Context, stateHash []byte) (*identity.LoginAttempt, error) {
	row, err := repository.queries.TakeLoginAttempt(ctx, stateHash)
	if err != nil {
		return nil, notFound(err, "login attempt")
	}
	var inviteLinkID *shared.InviteLinkID
	if row.InviteLinkID != nil {
		id := shared.InviteLinkID(row.InviteLinkID.String())
		inviteLinkID = &id
	}
	return repository.factory.Reconstruct(
		shared.LoginAttemptID(row.ID.String()),
		row.StateHash,
		row.Nonce,
		row.CodeVerifier,
		inviteLinkID,
		row.ExpiresAt.Time,
	), nil
}
