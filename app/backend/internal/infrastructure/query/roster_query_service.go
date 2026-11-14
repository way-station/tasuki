// Package query holds the query services that read from both PostgreSQL and MongoDB. Each
// database is read by its own package; this package only combines what they return
// (ADR-0011).
package query

import (
	"context"
	"slices"

	"github.com/way-station/tasuki/app/backend/internal/application/rosters"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres"
)

// RosterReader reads the roster from PostgreSQL. It is implemented by
// postgres.RosterReader.
type RosterReader interface {
	ListRosterMembers(ctx context.Context, teamID shared.TeamID) ([]rosters.RosterMemberRecord, error)
	GetRosterMember(ctx context.Context, teamID shared.TeamID, memberID shared.MemberID) (rosters.RosterMemberDetailRecord, error)
	ListPendingLinks(ctx context.Context, teamID shared.TeamID) ([]rosters.PendingLinkRecord, error)
	ListInvitedRaces(ctx context.Context, memberID shared.MemberID) ([]roster.InvitedRace, error)
	RacesOfTeam(ctx context.Context, teamID shared.TeamID, raceIDs []shared.RaceID) ([]postgres.RaceSummary, error)
}

// MemberLapReader reads a member's laps from MongoDB. It is implemented by
// mongodb.MemberLapReader.
type MemberLapReader interface {
	LapsByRace(ctx context.Context, memberID shared.MemberID) (map[shared.RaceID][]rosters.LapRecord, error)
}

// RosterQueryService implements rosters.RosterQueryService: the roster from PostgreSQL,
// and a member's laps from MongoDB.
type RosterQueryService struct {
	roster RosterReader
	laps   MemberLapReader
}

var _ rosters.RosterQueryService = (*RosterQueryService)(nil)

// NewRosterQueryService returns a RosterQueryService that reads through rosterReader and
// laps.
func NewRosterQueryService(rosterReader RosterReader, laps MemberLapReader) *RosterQueryService {
	return &RosterQueryService{roster: rosterReader, laps: laps}
}

// ListRosterMembers returns every roster member of the team, by name.
func (service *RosterQueryService) ListRosterMembers(ctx context.Context, teamID shared.TeamID) ([]rosters.RosterMemberRecord, error) {
	return service.roster.ListRosterMembers(ctx, teamID)
}

// GetRosterMember returns the member with their laps of each race that has some, newest
// race date first (docs/design/api.md). It returns an error wrapping shared.ErrNotFound
// when the team has no such member.
func (service *RosterQueryService) GetRosterMember(ctx context.Context, teamID shared.TeamID, memberID shared.MemberID) (rosters.RosterMemberDetailRecord, error) {
	record, err := service.roster.GetRosterMember(ctx, teamID, memberID)
	if err != nil {
		return rosters.RosterMemberDetailRecord{}, err
	}
	byRace, err := service.laps.LapsByRace(ctx, memberID)
	if err != nil {
		return rosters.RosterMemberDetailRecord{}, err
	}
	raceIDs := make([]shared.RaceID, 0, len(byRace))
	for raceID := range byRace {
		raceIDs = append(raceIDs, raceID)
	}
	slices.Sort(raceIDs)
	races, err := service.roster.RacesOfTeam(ctx, teamID, raceIDs)
	if err != nil {
		return rosters.RosterMemberDetailRecord{}, err
	}
	record.LapsByRace = make([]rosters.RaceLapsRecord, 0, len(races))
	for _, race := range races {
		record.LapsByRace = append(record.LapsByRace, rosters.RaceLapsRecord{RaceID: race.ID, RaceName: race.Name, Laps: byRace[race.ID]})
	}
	return record, nil
}

// ListPendingLinks returns the links that wait for approval, oldest first.
func (service *RosterQueryService) ListPendingLinks(ctx context.Context, teamID shared.TeamID) ([]rosters.PendingLinkRecord, error) {
	return service.roster.ListPendingLinks(ctx, teamID)
}

// ListInvitedRaces returns the races the member is invited to, with their dates.
func (service *RosterQueryService) ListInvitedRaces(ctx context.Context, memberID shared.MemberID) ([]roster.InvitedRace, error) {
	return service.roster.ListInvitedRaces(ctx, memberID)
}
