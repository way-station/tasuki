package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// RosterMemberRepository implements roster.MemberRepository with PostgreSQL. It runs in
// the transaction of a unit of work when there is one.
type RosterMemberRepository struct {
	queries *sqlcgen.Queries
	factory roster.MemberFactory
}

var _ roster.MemberRepository = (*RosterMemberRepository)(nil)

// NewRosterMemberRepository returns a RosterMemberRepository that uses client.
func NewRosterMemberRepository(client *Client, factory roster.MemberFactory) *RosterMemberRepository {
	return &RosterMemberRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// FindByID returns the roster member with id.
func (repository *RosterMemberRepository) FindByID(ctx context.Context, id shared.MemberID) (*roster.Member, error) {
	memberID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	row, err := queriesFor(ctx, repository.queries).GetRosterMember(ctx, memberID)
	if err != nil {
		return nil, notFound(err, "roster member")
	}
	return repository.factory.Reconstruct(id, shared.TeamID(row.TeamID.String()), row.Name, timeOrNil(row.LeftAt)), nil
}

// Save inserts a new member, or stores whether a stored member has left.
func (repository *RosterMemberRepository) Save(ctx context.Context, member *roster.Member) error {
	memberID, err := parseID(member.ID())
	if err != nil {
		return err
	}
	teamID, err := parseID(member.TeamID())
	if err != nil {
		return err
	}
	var leftAt pgtype.Timestamptz
	if at, ok := member.LeftAt(); ok {
		leftAt = timestamptz(at)
	}
	return queriesFor(ctx, repository.queries).SaveRosterMember(ctx, sqlcgen.SaveRosterMemberParams{
		ID:     memberID,
		TeamID: teamID,
		Name:   member.Name(),
		LeftAt: leftAt,
	})
}
