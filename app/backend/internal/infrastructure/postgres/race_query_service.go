package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/way-station/tasuki/app/backend/internal/application/races"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// RaceQueryService implements races.RaceQueryService with PostgreSQL.
type RaceQueryService struct {
	queries *sqlcgen.Queries
}

var _ races.RaceQueryService = (*RaceQueryService)(nil)

// NewRaceQueryService returns a RaceQueryService that uses client.
func NewRaceQueryService(client *Client) *RaceQueryService {
	return &RaceQueryService{queries: sqlcgen.New(client.Pool())}
}

// ListRaces returns every race of the team, newest registered first.
func (queryService *RaceQueryService) ListRaces(ctx context.Context, teamID shared.TeamID, rosterMemberID *shared.MemberID) ([]races.RaceSummaryRecord, error) {
	team, err := parseID(teamID)
	if err != nil {
		return nil, err
	}
	rosterMember, err := parseOptionalID(rosterMemberID)
	if err != nil {
		return nil, err
	}
	rows, err := queryService.queries.ListRaceSummaries(ctx, sqlcgen.ListRaceSummariesParams{RosterMemberID: rosterMember, TeamID: team})
	if err != nil {
		return nil, err
	}

	records := make([]races.RaceSummaryRecord, 0, len(rows))
	for _, row := range rows {
		params, err := decodeRaceDescription(row.Description)
		if err != nil {
			return nil, err
		}
		answer, err := answerOf(row.Invited, row.Attendance)
		if err != nil {
			return nil, err
		}
		records = append(records, races.RaceSummaryRecord{
			ID:                    shared.RaceID(row.ID.String()),
			Name:                  row.Name,
			RaceDate:              params.RaceDate,
			Answer:                answer,
			RunningOrderConfirmed: row.RunningOrderConfirmed,
		})
	}
	return records, nil
}

// FindRace returns the race of the team.
func (queryService *RaceQueryService) FindRace(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID, rosterMemberID *shared.MemberID) (races.RaceRecord, error) {
	team, err := parseID(teamID)
	if err != nil {
		return races.RaceRecord{}, err
	}
	id, err := parseID(raceID)
	if err != nil {
		return races.RaceRecord{}, err
	}
	rosterMember, err := parseOptionalID(rosterMemberID)
	if err != nil {
		return races.RaceRecord{}, err
	}
	row, err := queryService.queries.GetRaceView(ctx, sqlcgen.GetRaceViewParams{RosterMemberID: rosterMember, TeamID: team, RaceID: id})
	if err != nil {
		return races.RaceRecord{}, notFound(err, "race")
	}
	params, err := decodeRaceDescription(row.Description)
	if err != nil {
		return races.RaceRecord{}, err
	}
	answer, err := answerOf(row.Invited, row.Attendance)
	if err != nil {
		return races.RaceRecord{}, err
	}
	return races.RaceRecord{
		ID:                    raceID,
		Name:                  row.Name,
		Version:               int(row.Version),
		Description:           params,
		Answer:                answer,
		RunningOrderConfirmed: row.RunningOrderConfirmed,
		RunsInConfirmedOrder:  row.RunsInConfirmedOrder,
	}, nil
}

// answerOf converts an invitation and its attendance row to the answer. An invitation
// without an attendance row has not been answered.
func answerOf(invited bool, attendance pgtype.Text) (race.Answer, error) {
	if !invited {
		return race.NotInvited, nil
	}
	if !attendance.Valid {
		return race.NotAnswered, nil
	}
	switch attendance.String {
	case "attending":
		return race.Attending, nil
	case "not_attending":
		return race.NotAttending, nil
	case "undecided":
		return race.Undecided, nil
	}
	return 0, fmt.Errorf("unknown attendance status %q", attendance.String)
}
