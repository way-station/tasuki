package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/way-station/tasuki/app/backend/internal/application/runningorders"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// RunningOrderQueryService implements runningorders.RunningOrderQueryService with
// PostgreSQL.
type RunningOrderQueryService struct {
	queries *sqlcgen.Queries
}

var _ runningorders.RunningOrderQueryService = (*RunningOrderQueryService)(nil)

// NewRunningOrderQueryService returns a RunningOrderQueryService that uses client.
func NewRunningOrderQueryService(client *Client) *RunningOrderQueryService {
	return &RunningOrderQueryService{queries: sqlcgen.New(client.Pool())}
}

// ListRunningOrders returns the race with the answer of rosterMemberID, and the saved
// running orders of its entries with their legs, runners, and average times.
func (queryService *RunningOrderQueryService) ListRunningOrders(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID, rosterMemberID *shared.MemberID) (runningorders.RaceRunningOrdersRecord, error) {
	team, err := parseID(teamID)
	if err != nil {
		return runningorders.RaceRunningOrdersRecord{}, err
	}
	raceUUID, err := parseID(raceID)
	if err != nil {
		return runningorders.RaceRunningOrdersRecord{}, err
	}
	rosterMember, err := parseOptionalID(rosterMemberID)
	if err != nil {
		return runningorders.RaceRunningOrdersRecord{}, err
	}
	raceRow, err := queryService.queries.GetRunningOrderRace(ctx, sqlcgen.GetRunningOrderRaceParams{RosterMemberID: rosterMember, RaceID: raceUUID, TeamID: team})
	if err != nil {
		return runningorders.RaceRunningOrdersRecord{}, notFound(err, "race")
	}
	params, err := decodeRaceDescription(raceRow.Description)
	if err != nil {
		return runningorders.RaceRunningOrdersRecord{}, err
	}
	answer, err := runningOrderViewerAnswer(raceRow.Invited, raceRow.Attendance)
	if err != nil {
		return runningorders.RaceRunningOrdersRecord{}, err
	}
	orders, err := queryService.runningOrders(ctx, raceUUID)
	if err != nil {
		return runningorders.RaceRunningOrdersRecord{}, err
	}
	return runningorders.RaceRunningOrdersRecord{Description: params, Answer: answer, RunningOrders: orders}, nil
}

func (queryService *RunningOrderQueryService) runningOrders(ctx context.Context, raceID uuid.UUID) ([]runningorders.RunningOrderRecord, error) {
	rows, err := queryService.queries.ListRunningOrderViews(ctx, raceID)
	if err != nil {
		return nil, err
	}
	legRows, err := queryService.queries.ListRunningOrderLegViews(ctx, raceID)
	if err != nil {
		return nil, err
	}
	legs := map[uuid.UUID][]runningorders.LegRecord{}
	for _, legRow := range legRows {
		leg := runningorders.LegRecord{Number: int(legRow.LegNumber)}
		if legRow.RunnerID != nil {
			leg.Runner = &runningorders.RunnerRecord{
				MemberID:       shared.MemberID(legRow.RunnerID.String()),
				Name:           legRow.RunnerName.String,
				AverageSeconds: int(legRow.RunnerAverageSeconds.Int32),
			}
		}
		legs[legRow.RunningOrderID] = append(legs[legRow.RunningOrderID], leg)
	}
	records := make([]runningorders.RunningOrderRecord, 0, len(rows))
	for _, row := range rows {
		records = append(records, runningorders.RunningOrderRecord{
			EntryID:   shared.EntryID(row.EntryID.String()),
			EntryName: row.EntryName,
			Status:    runningorder.Status(row.Status),
			Version:   int(row.Version),
			Legs:      legs[row.ID],
		})
	}
	return records, nil
}

// runningOrderViewerAnswer returns the answer as the visibility policy sees it.
func runningOrderViewerAnswer(invited bool, attendance pgtype.Text) (race.Answer, error) {
	switch {
	case !invited:
		return race.NotInvited, nil
	case !attendance.Valid:
		return race.NotAnswered, nil
	}
	switch attendance.String {
	case "attending":
		return race.Attending, nil
	case "undecided":
		return race.Undecided, nil
	case "not_attending":
		return race.NotAttending, nil
	}
	return race.NotInvited, fmt.Errorf("unknown attendance status %q", attendance.String)
}
