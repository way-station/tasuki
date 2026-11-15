package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/way-station/tasuki/app/backend/internal/application/participations"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// EntryQueryService implements participations.EntryQueryService with PostgreSQL.
type EntryQueryService struct {
	queries *sqlcgen.Queries
}

var _ participations.EntryQueryService = (*EntryQueryService)(nil)

// NewEntryQueryService returns an EntryQueryService that uses client.
func NewEntryQueryService(client *Client) *EntryQueryService {
	return &EntryQueryService{queries: sqlcgen.New(client.Pool())}
}

// ListEntries returns the entries of the race, oldest first, with their members in the
// order they were added.
func (queryService *EntryQueryService) ListEntries(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]participations.EntryRecord, error) {
	_, race, err := teamRace(ctx, queryService.queries, teamID, raceID)
	if err != nil {
		return nil, err
	}
	rows, err := queryService.queries.ListEntryViews(ctx, race)
	if err != nil {
		return nil, err
	}
	memberRows, err := queryService.queries.ListEntryMemberViews(ctx, race)
	if err != nil {
		return nil, err
	}
	members := map[uuid.UUID][]participations.EntryMemberRecord{}
	for _, memberRow := range memberRows {
		members[memberRow.EntryID] = append(members[memberRow.EntryID], participations.EntryMemberRecord{MemberID: shared.MemberID(memberRow.ID.String()), Name: memberRow.Name})
	}
	records := make([]participations.EntryRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, participations.EntryRecord{ID: shared.EntryID(row.ID.String()), Name: row.Name, Members: members[row.ID]})
	}
	return records, nil
}
