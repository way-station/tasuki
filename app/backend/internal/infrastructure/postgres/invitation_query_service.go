package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/way-station/tasuki/app/backend/internal/application/participations"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// InvitationQueryService implements participations.InvitationQueryService with PostgreSQL.
type InvitationQueryService struct {
	queries *sqlcgen.Queries
}

var _ participations.InvitationQueryService = (*InvitationQueryService)(nil)

// NewInvitationQueryService returns an InvitationQueryService that uses client.
func NewInvitationQueryService(client *Client) *InvitationQueryService {
	return &InvitationQueryService{queries: sqlcgen.New(client.Pool())}
}

// ListInvitations returns every roster member of the team who has not left, in the order
// they were added to the roster.
func (queryService *InvitationQueryService) ListInvitations(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID) ([]participations.InvitationRecord, error) {
	team, race, err := teamRace(ctx, queryService.queries, teamID, raceID)
	if err != nil {
		return nil, err
	}
	rows, err := queryService.queries.ListInvitationViews(ctx, sqlcgen.ListInvitationViewsParams{RaceID: race, TeamID: team})
	if err != nil {
		return nil, err
	}
	records := make([]participations.InvitationRecord, 0, len(rows))
	for _, row := range rows {
		record := participations.InvitationRecord{MemberID: shared.MemberID(row.ID.String()), Name: row.Name, Invited: row.Invited}
		if row.Attendance.Valid {
			status := participation.AttendanceStatus(row.Attendance.String)
			record.Attendance = &status
		}
		records = append(records, record)
	}
	return records, nil
}

// teamRace returns the column values of the team and the race, or an error wrapping
// shared.ErrNotFound when the team has no such race.
func teamRace(ctx context.Context, queries *sqlcgen.Queries, teamID shared.TeamID, raceID shared.RaceID) (team, race uuid.UUID, err error) {
	team, err = parseID(teamID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	race, err = parseID(raceID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	exists, err := queries.TeamHasRace(ctx, sqlcgen.TeamHasRaceParams{RaceID: race, TeamID: team})
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	if !exists {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	return team, race, nil
}
