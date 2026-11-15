package postgres

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// AttendanceRepository implements participation.AttendanceRepository with PostgreSQL. It
// runs in the transaction of a unit of work when there is one.
type AttendanceRepository struct {
	queries *sqlcgen.Queries
	factory participation.AttendanceFactory
}

var _ participation.AttendanceRepository = (*AttendanceRepository)(nil)

// NewAttendanceRepository returns an AttendanceRepository that uses client.
func NewAttendanceRepository(client *Client, factory participation.AttendanceFactory) *AttendanceRepository {
	return &AttendanceRepository{queries: sqlcgen.New(client.Pool()), factory: factory}
}

// Find returns the member's answer for the race.
func (repository *AttendanceRepository) Find(ctx context.Context, raceID shared.RaceID, memberID shared.MemberID) (*participation.Attendance, error) {
	race, err := parseID(raceID)
	if err != nil {
		return nil, err
	}
	member, err := parseID(memberID)
	if err != nil {
		return nil, err
	}
	row, err := queriesFor(ctx, repository.queries).GetAttendance(ctx, sqlcgen.GetAttendanceParams{RaceID: race, RosterMemberID: member})
	if err != nil {
		return nil, notFound(err, "attendance")
	}
	return repository.factory.Reconstruct(raceID, memberID, participation.AttendanceStatus(row.Status)), nil
}

// Save stores the first answer, or replaces the stored one.
func (repository *AttendanceRepository) Save(ctx context.Context, attendance *participation.Attendance) error {
	race, err := parseID(attendance.RaceID())
	if err != nil {
		return err
	}
	member, err := parseID(attendance.MemberID())
	if err != nil {
		return err
	}
	return queriesFor(ctx, repository.queries).SaveAttendance(ctx, sqlcgen.SaveAttendanceParams{
		RaceID:         race,
		RosterMemberID: member,
		Status:         string(attendance.Status()),
	})
}
