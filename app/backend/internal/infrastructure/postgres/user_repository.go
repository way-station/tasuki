package postgres

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// UserRepository implements identity.UserRepository with PostgreSQL.
type UserRepository struct {
	queries *sqlcgen.Queries
	factory identity.UserFactory
}

var _ identity.UserRepository = (*UserRepository)(nil)

// NewUserRepository returns an UserRepository that uses client.
func NewUserRepository(client *Client, factory identity.UserFactory) *UserRepository {
	return &UserRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// FindByID returns the user with id.
func (repository *UserRepository) FindByID(ctx context.Context, id shared.UserID) (*identity.User, error) {
	userID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	row, err := repository.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, notFound(err, "user")
	}
	return repository.factory.Reconstruct(shared.UserID(row.ID.String()), row.LineUserID, row.DisplayName), nil
}

// FindByLineUserID returns the user of a LINE user.
func (repository *UserRepository) FindByLineUserID(ctx context.Context, lineUserID string) (*identity.User, error) {
	row, err := repository.queries.GetUserByLineUserID(ctx, lineUserID)
	if err != nil {
		return nil, notFound(err, "user")
	}
	return repository.factory.Reconstruct(shared.UserID(row.ID.String()), row.LineUserID, row.DisplayName), nil
}

// Save creates the user, or updates its display name.
func (repository *UserRepository) Save(ctx context.Context, user *identity.User) error {
	userID, err := parseID(user.ID())
	if err != nil {
		return err
	}
	return repository.queries.SaveUser(ctx, sqlcgen.SaveUserParams{
		ID:          userID,
		LineUserID:  user.LineUserID(),
		DisplayName: user.DisplayName(),
	})
}
