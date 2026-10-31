package postgres

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// AccountRepository implements identity.AccountRepository with PostgreSQL.
type AccountRepository struct {
	queries *sqlcgen.Queries
	factory identity.AccountFactory
}

var _ identity.AccountRepository = (*AccountRepository)(nil)

// NewAccountRepository returns an AccountRepository that uses client.
func NewAccountRepository(client *Client, factory identity.AccountFactory) *AccountRepository {
	return &AccountRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// FindByID returns the account with id.
func (repository *AccountRepository) FindByID(ctx context.Context, id shared.AccountID) (*identity.Account, error) {
	accountID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	row, err := repository.queries.GetAccountByID(ctx, accountID)
	if err != nil {
		return nil, notFound(err, "account")
	}
	return repository.factory.Reconstruct(shared.AccountID(row.ID.String()), row.LineUserID, row.DisplayName), nil
}

// FindByLineUserID returns the account of a LINE user.
func (repository *AccountRepository) FindByLineUserID(ctx context.Context, lineUserID string) (*identity.Account, error) {
	row, err := repository.queries.GetAccountByLineUserID(ctx, lineUserID)
	if err != nil {
		return nil, notFound(err, "account")
	}
	return repository.factory.Reconstruct(shared.AccountID(row.ID.String()), row.LineUserID, row.DisplayName), nil
}

// Save creates the account, or updates its display name.
func (repository *AccountRepository) Save(ctx context.Context, account *identity.Account) error {
	accountID, err := parseID(account.ID())
	if err != nil {
		return err
	}
	return repository.queries.SaveAccount(ctx, sqlcgen.SaveAccountParams{
		ID:          accountID,
		LineUserID:  account.LineUserID(),
		DisplayName: account.DisplayName(),
	})
}
