package postgres

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// InviteLinkQueryService implements auth.InviteLinkQueryService with PostgreSQL. It runs in
// the transaction of a unit of work when there is one.
type InviteLinkQueryService struct {
	queries *sqlcgen.Queries
}

var _ auth.InviteLinkQueryService = (*InviteLinkQueryService)(nil)

// NewInviteLinkQueryService returns an InviteLinkQueryService that uses client.
func NewInviteLinkQueryService(client *Client) *InviteLinkQueryService {
	return &InviteLinkQueryService{queries: sqlcgen.New(client.Pool())}
}

// FindByTokenHash returns the invite link whose token has hash, with its team name.
func (service *InviteLinkQueryService) FindByTokenHash(ctx context.Context, hash []byte) (auth.InviteLinkRecord, error) {
	row, err := queriesFor(ctx, service.queries).GetInviteLinkWithTeamByTokenHash(ctx, hash)
	if err != nil {
		return auth.InviteLinkRecord{}, notFound(err, "invite link")
	}
	return inviteLinkRecord(sqlcgen.GetInviteLinkWithTeamByIDRow(row)), nil
}

// FindByID returns the invite link with its team name. In a unit of work, the link stays
// locked until the transaction ends.
func (service *InviteLinkQueryService) FindByID(ctx context.Context, id shared.InviteLinkID) (auth.InviteLinkRecord, error) {
	linkID, err := parseID(id)
	if err != nil {
		return auth.InviteLinkRecord{}, err
	}
	row, err := queriesFor(ctx, service.queries).GetInviteLinkWithTeamByID(ctx, linkID)
	if err != nil {
		return auth.InviteLinkRecord{}, notFound(err, "invite link")
	}
	return inviteLinkRecord(row), nil
}

func inviteLinkRecord(row sqlcgen.GetInviteLinkWithTeamByIDRow) auth.InviteLinkRecord {
	return auth.InviteLinkRecord{
		ID:        shared.InviteLinkID(row.ID.String()),
		MemberID:  shared.MemberID(row.RosterMemberID.String()),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt.Time,
		UsedAt:    timeOrNil(row.UsedAt),
		RevokedAt: timeOrNil(row.RevokedAt),
		TeamName:  row.TeamName,
	}
}
