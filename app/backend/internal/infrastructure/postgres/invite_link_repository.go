package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// InviteLinkRepository implements identity.InviteLinkRepository with PostgreSQL. It runs in
// the transaction of a unit of work when there is one.
type InviteLinkRepository struct {
	queries *sqlcgen.Queries
	factory identity.InviteLinkFactory
}

var _ identity.InviteLinkRepository = (*InviteLinkRepository)(nil)

// NewInviteLinkRepository returns an InviteLinkRepository that uses client.
func NewInviteLinkRepository(client *Client, factory identity.InviteLinkFactory) *InviteLinkRepository {
	return &InviteLinkRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// FindByTokenHash returns the link whose token has hash, whatever its state.
func (repository *InviteLinkRepository) FindByTokenHash(ctx context.Context, hash []byte) (*identity.InviteLink, error) {
	row, err := queriesFor(ctx, repository.queries).GetInviteLinkByTokenHash(ctx, hash)
	if err != nil {
		return nil, notFound(err, "invite link")
	}
	return repository.reconstruct(sqlcgen.InviteLink{ID: row.ID, RosterMemberID: row.RosterMemberID, TokenHash: row.TokenHash, ExpiresAt: row.ExpiresAt, UsedAt: row.UsedAt, RevokedAt: row.RevokedAt}), nil
}

// FindOpenByMember returns the member's link that is neither used nor replaced.
func (repository *InviteLinkRepository) FindOpenByMember(ctx context.Context, memberID shared.MemberID) (*identity.InviteLink, error) {
	member, err := parseID(memberID)
	if err != nil {
		return nil, err
	}
	row, err := queriesFor(ctx, repository.queries).GetOpenInviteLinkOfMember(ctx, member)
	if err != nil {
		return nil, notFound(err, "open invite link")
	}
	return repository.reconstruct(sqlcgen.InviteLink{ID: row.ID, RosterMemberID: row.RosterMemberID, TokenHash: row.TokenHash, ExpiresAt: row.ExpiresAt, UsedAt: row.UsedAt, RevokedAt: row.RevokedAt}), nil
}

// Save inserts a new link, or records that it was used or replaced.
func (repository *InviteLinkRepository) Save(ctx context.Context, link *identity.InviteLink) error {
	linkID, err := parseID(link.ID())
	if err != nil {
		return err
	}
	memberID, err := parseID(link.MemberID())
	if err != nil {
		return err
	}
	var usedAt, revokedAt pgtype.Timestamptz
	if at, ok := link.UsedAt(); ok {
		usedAt = timestamptz(at)
	}
	if at, ok := link.RevokedAt(); ok {
		revokedAt = timestamptz(at)
	}
	return queriesFor(ctx, repository.queries).SaveInviteLink(ctx, sqlcgen.SaveInviteLinkParams{
		ID:             linkID,
		RosterMemberID: memberID,
		TokenHash:      link.TokenHash(),
		ExpiresAt:      timestamptz(link.ExpiresAt()),
		UsedAt:         usedAt,
		RevokedAt:      revokedAt,
	})
}

func (repository *InviteLinkRepository) reconstruct(row sqlcgen.InviteLink) *identity.InviteLink {
	return repository.factory.Reconstruct(shared.InviteLinkID(row.ID.String()), shared.MemberID(row.RosterMemberID.String()), row.TokenHash, row.ExpiresAt.Time, timeOrNil(row.UsedAt), timeOrNil(row.RevokedAt))
}
