package postgres

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/application/laps"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// ProfileQueryService implements laps.ProfileQueryService with roster_members.
type ProfileQueryService struct {
	queries *sqlcgen.Queries
}

var _ laps.ProfileQueryService = (*ProfileQueryService)(nil)

// NewProfileQueryService returns a ProfileQueryService that uses client.
func NewProfileQueryService(client *Client) *ProfileQueryService {
	return &ProfileQueryService{queries: sqlcgen.New(client.Pool())}
}

// GetProfile returns the member's name, average time, and lap count. It returns an error
// wrapping shared.ErrNotFound when there is no such member.
func (service *ProfileQueryService) GetProfile(ctx context.Context, memberID shared.MemberID) (laps.ProfileRecord, error) {
	member, err := parseID(memberID)
	if err != nil {
		return laps.ProfileRecord{}, err
	}
	row, err := queriesFor(ctx, service.queries).GetProfile(ctx, member)
	if err != nil {
		return laps.ProfileRecord{}, notFound(err, "roster member")
	}
	average, err := shared.PaceFromSeconds(int(row.AverageSeconds))
	if err != nil {
		return laps.ProfileRecord{}, fmt.Errorf("roster member %s has an invalid stored average: %w", memberID, err)
	}
	return laps.ProfileRecord{Name: row.Name, Average: average, LapCount: int(row.LapCount)}, nil
}
