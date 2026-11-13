package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/way-station/tasuki/app/backend/internal/application/rosters"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// RosterReader reads what the roster query service needs from PostgreSQL (ADR-0011). The
// laps are read from MongoDB by the query service.
type RosterReader struct {
	queries *sqlcgen.Queries
}

// NewRosterReader returns a RosterReader that uses client.
func NewRosterReader(client *Client) *RosterReader {
	return &RosterReader{queries: sqlcgen.New(client.Pool())}
}

// RaceSummary is the name and date of a race.
type RaceSummary struct {
	ID       shared.RaceID
	Name     string
	RaceDate string
}

// ListRosterMembers returns every roster member of the team, by name, with their links
// and their invite link that is neither used nor replaced.
func (reader *RosterReader) ListRosterMembers(ctx context.Context, teamID shared.TeamID) ([]rosters.RosterMemberRecord, error) {
	team, err := parseID(teamID)
	if err != nil {
		return nil, err
	}
	queries := queriesFor(ctx, reader.queries)
	memberRows, err := queries.ListRosterMembersOfTeam(ctx, team)
	if err != nil {
		return nil, err
	}
	linkRows, err := queries.ListMemberLinksOfTeam(ctx, team)
	if err != nil {
		return nil, err
	}
	inviteRows, err := queries.ListOpenInviteLinksOfTeam(ctx, team)
	if err != nil {
		return nil, err
	}
	links := make(map[uuid.UUID][]rosters.LinkRecord, len(linkRows))
	for _, row := range linkRows {
		links[row.RosterMemberID] = append(links[row.RosterMemberID], rosters.LinkRecord{
			ID:           shared.MemberLinkID(row.ID.String()),
			UserID:       shared.UserID(row.UserID.String()),
			InviteLinkID: shared.InviteLinkID(row.InviteLinkID.String()),
			CreatedAt:    row.CreatedAt.Time,
			ApprovedAt:   timeOrNil(row.ApprovedAt),
			RejectedAt:   timeOrNil(row.RejectedAt),
			UnlinkedAt:   timeOrNil(row.UnlinkedAt),
		})
	}
	invites := make(map[uuid.UUID]*rosters.InviteLinkRecord, len(inviteRows))
	for _, row := range inviteRows {
		invites[row.RosterMemberID] = &rosters.InviteLinkRecord{ID: shared.InviteLinkID(row.ID.String()), ExpiresAt: row.ExpiresAt.Time}
	}
	members := make([]rosters.RosterMemberRecord, 0, len(memberRows))
	for _, row := range memberRows {
		members = append(members, rosters.RosterMemberRecord{
			MemberID:   shared.MemberID(row.ID.String()),
			Name:       row.Name,
			Left:       row.LeftAt.Valid,
			Links:      links[row.ID],
			InviteLink: invites[row.ID],
		})
	}
	return members, nil
}

// GetRosterMember returns the member of the team with their average time, without laps. It
// returns an error wrapping shared.ErrNotFound when the team has no such member.
func (reader *RosterReader) GetRosterMember(ctx context.Context, teamID shared.TeamID, memberID shared.MemberID) (rosters.RosterMemberDetailRecord, error) {
	team, err := parseID(teamID)
	if err != nil {
		return rosters.RosterMemberDetailRecord{}, err
	}
	member, err := parseID(memberID)
	if err != nil {
		return rosters.RosterMemberDetailRecord{}, err
	}
	row, err := queriesFor(ctx, reader.queries).GetRosterMemberOfTeam(ctx, sqlcgen.GetRosterMemberOfTeamParams{ID: member, TeamID: team})
	if err != nil {
		return rosters.RosterMemberDetailRecord{}, notFound(err, "roster member")
	}
	average, err := shared.PaceFromSeconds(int(row.AverageSeconds))
	if err != nil {
		return rosters.RosterMemberDetailRecord{}, fmt.Errorf("roster member %s has an invalid stored average: %w", memberID, err)
	}
	return rosters.RosterMemberDetailRecord{MemberID: memberID, Name: row.Name, Left: row.LeftAt.Valid, Average: average, LapCount: int(row.LapCount)}, nil
}

// ListPendingLinks returns the links of the team's members that wait for approval, oldest
// first.
func (reader *RosterReader) ListPendingLinks(ctx context.Context, teamID shared.TeamID) ([]rosters.PendingLinkRecord, error) {
	team, err := parseID(teamID)
	if err != nil {
		return nil, err
	}
	rows, err := queriesFor(ctx, reader.queries).ListPendingLinksOfTeam(ctx, team)
	if err != nil {
		return nil, err
	}
	links := make([]rosters.PendingLinkRecord, 0, len(rows))
	for _, row := range rows {
		links = append(links, rosters.PendingLinkRecord{
			LinkID:          shared.MemberLinkID(row.ID.String()),
			LineDisplayName: row.DisplayName,
			MemberID:        shared.MemberID(row.RosterMemberID.String()),
			MemberName:      row.RosterMemberName,
			CreatedAt:       row.CreatedAt.Time,
		})
	}
	return links, nil
}

// ListInvitedRaces returns the races the member is invited to, with their dates.
func (reader *RosterReader) ListInvitedRaces(ctx context.Context, memberID shared.MemberID) ([]roster.InvitedRace, error) {
	member, err := parseID(memberID)
	if err != nil {
		return nil, err
	}
	rows, err := queriesFor(ctx, reader.queries).ListInvitedRaces(ctx, member)
	if err != nil {
		return nil, err
	}
	races := make([]roster.InvitedRace, 0, len(rows))
	for _, row := range rows {
		raceDate, err := shared.ParseDate(row.RaceDate)
		if err != nil {
			return nil, fmt.Errorf("race %s has an invalid stored date: %w", row.ID, err)
		}
		races = append(races, roster.InvitedRace{RaceID: shared.RaceID(row.ID.String()), RaceDate: raceDate})
	}
	return races, nil
}

// RacesOfTeam returns the names and dates of the team's races among raceIDs, newest race
// date first. Races of other teams are left out.
func (reader *RosterReader) RacesOfTeam(ctx context.Context, teamID shared.TeamID, raceIDs []shared.RaceID) ([]RaceSummary, error) {
	team, err := parseID(teamID)
	if err != nil {
		return nil, err
	}
	ids := make([]uuid.UUID, 0, len(raceIDs))
	for _, raceID := range raceIDs {
		id, err := parseID(raceID)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	rows, err := queriesFor(ctx, reader.queries).ListRacesOfTeamByID(ctx, sqlcgen.ListRacesOfTeamByIDParams{TeamID: team, Ids: ids})
	if err != nil {
		return nil, err
	}
	races := make([]RaceSummary, 0, len(rows))
	for _, row := range rows {
		races = append(races, RaceSummary{ID: shared.RaceID(row.ID.String()), Name: row.Name, RaceDate: row.RaceDate})
	}
	return races, nil
}
