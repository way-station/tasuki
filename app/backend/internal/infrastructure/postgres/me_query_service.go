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

// FindMe returns the user and its member links, oldest first.
func (queryService *MeQueryService) FindMe(ctx context.Context, userID shared.UserID) (auth.MeRecord, error) {
	id, err := parseID(userID)
	if err != nil {
		return auth.MeRecord{}, err
	}
	user, err := queryService.queries.GetMeUser(ctx, id)
	if err != nil {
		return auth.MeRecord{}, notFound(err, "user")
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
	return auth.MeRecord{LineUserID: user.LineUserID, DisplayName: user.DisplayName, Links: links}, nil
}
