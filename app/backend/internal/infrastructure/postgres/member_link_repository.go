package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// MemberLinkRepository implements identity.MemberLinkRepository with PostgreSQL. It runs in
// the transaction of a unit of work when there is one.
type MemberLinkRepository struct {
	queries *sqlcgen.Queries
	factory identity.MemberLinkFactory
}

var _ identity.MemberLinkRepository = (*MemberLinkRepository)(nil)

// NewMemberLinkRepository returns a MemberLinkRepository that uses client.
func NewMemberLinkRepository(client *Client, factory identity.MemberLinkFactory) *MemberLinkRepository {
	return &MemberLinkRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// FindByID returns the link with id.
func (repository *MemberLinkRepository) FindByID(ctx context.Context, id shared.MemberLinkID) (*identity.MemberLink, error) {
	linkID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	row, err := queriesFor(ctx, repository.queries).GetMemberLink(ctx, linkID)
	if err != nil {
		return nil, notFound(err, "member link")
	}
	return repository.reconstruct(row), nil
}

// ListByUser returns the user's links, oldest first.
func (repository *MemberLinkRepository) ListByUser(ctx context.Context, userID shared.UserID) ([]*identity.MemberLink, error) {
	user, err := parseID(userID)
	if err != nil {
		return nil, err
	}
	rows, err := queriesFor(ctx, repository.queries).ListMemberLinksByUser(ctx, user)
	if err != nil {
		return nil, err
	}
	links := make([]*identity.MemberLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, repository.reconstruct(row))
	}
	return links, nil
}

// ListByMember returns the roster member's links, oldest first.
func (repository *MemberLinkRepository) ListByMember(ctx context.Context, memberID shared.MemberID) ([]*identity.MemberLink, error) {
	member, err := parseID(memberID)
	if err != nil {
		return nil, err
	}
	rows, err := queriesFor(ctx, repository.queries).ListMemberLinksByMember(ctx, member)
	if err != nil {
		return nil, err
	}
	links := make([]*identity.MemberLink, 0, len(rows))
	for _, row := range rows {
		links = append(links, repository.reconstruct(row))
	}
	return links, nil
}

// Save inserts a new link, or records that it was approved, rejected, or unlinked.
func (repository *MemberLinkRepository) Save(ctx context.Context, link *identity.MemberLink) error {
	linkID, err := parseID(link.ID())
	if err != nil {
		return err
	}
	userID, err := parseID(link.UserID())
	if err != nil {
		return err
	}
	memberID, err := parseID(link.MemberID())
	if err != nil {
		return err
	}
	inviteLinkID, err := parseID(link.InviteLinkID())
	if err != nil {
		return err
	}
	return queriesFor(ctx, repository.queries).SaveMemberLink(ctx, sqlcgen.SaveMemberLinkParams{
		ID:             linkID,
		UserID:         userID,
		RosterMemberID: memberID,
		InviteLinkID:   inviteLinkID,
		CreatedAt:      timestamptz(link.CreatedAt()),
		ApprovedAt:     timestamptzOf(link.ApprovedAt()),
		RejectedAt:     timestamptzOf(link.RejectedAt()),
		UnlinkedAt:     timestamptzOf(link.UnlinkedAt()),
	})
}

func (repository *MemberLinkRepository) reconstruct(row sqlcgen.MemberLink) *identity.MemberLink {
	return repository.factory.Reconstruct(
		shared.MemberLinkID(row.ID.String()),
		shared.UserID(row.UserID.String()),
		shared.MemberID(row.RosterMemberID.String()),
		shared.InviteLinkID(row.InviteLinkID.String()),
		row.CreatedAt.Time,
		timeOrNil(row.ApprovedAt),
		timeOrNil(row.RejectedAt),
		timeOrNil(row.UnlinkedAt),
	)
}

// timestamptzOf converts a time that may not be set, as the domain returns it, to a
// nullable column value.
func timestamptzOf(moment time.Time, set bool) pgtype.Timestamptz {
	if !set {
		return pgtype.Timestamptz{}
	}
	return timestamptz(moment)
}
