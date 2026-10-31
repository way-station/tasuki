package postgres

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// MeQueryService implements auth.MeQueryService with PostgreSQL.
type MeQueryService struct {
	queries *sqlcgen.Queries
}

var _ auth.MeQueryService = (*MeQueryService)(nil)

// NewMeQueryService returns a MeQueryService that uses client.
func NewMeQueryService(client *Client) *MeQueryService {
	return &MeQueryService{queries: sqlcgen.New(client.Pool())}
}

// FindMe returns the account and its member links, oldest first.
func (queryService *MeQueryService) FindMe(ctx context.Context, accountID shared.AccountID) (auth.MeRecord, error) {
	id, err := parseID(accountID)
	if err != nil {
		return auth.MeRecord{}, err
	}
	account, err := queryService.queries.GetMeAccount(ctx, id)
	if err != nil {
		return auth.MeRecord{}, notFound(err, "account")
	}
	rows, err := queryService.queries.ListMeLinks(ctx, id)
	if err != nil {
		return auth.MeRecord{}, err
	}

	links := make([]auth.MeLinkRecord, 0, len(rows))
	for _, row := range rows {
		links = append(links, auth.MeLinkRecord{
			ID:               shared.MemberLinkID(row.ID.String()),
			RosterMemberID:   shared.MemberID(row.RosterMemberID.String()),
			RosterMemberName: row.RosterMemberName,
			InviteLinkID:     shared.InviteLinkID(row.InviteLinkID.String()),
			CreatedAt:        row.CreatedAt.Time,
			ApprovedAt:       timeOrNil(row.ApprovedAt),
			RejectedAt:       timeOrNil(row.RejectedAt),
			UnlinkedAt:       timeOrNil(row.UnlinkedAt),
		})
	}
	return auth.MeRecord{LineUserID: account.LineUserID, DisplayName: account.DisplayName, Links: links}, nil
}
