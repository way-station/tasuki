package races

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// RaceQueryService reads races as the race screens show them, together with what a
// roster member answered. It is implemented in the infrastructure layer.
type RaceQueryService interface {
	// ListRaces returns every race of the team, newest registered first. rosterMemberID is
	// whose answers to read; with nil, every answer is race.NotInvited.
	ListRaces(ctx context.Context, teamID shared.TeamID, rosterMemberID *shared.MemberID) ([]RaceSummaryRecord, error)
	// FindRace returns the race of the team. It returns an error wrapping
	// shared.ErrNotFound when there is none. rosterMemberID is as in ListRaces.
	FindRace(ctx context.Context, teamID shared.TeamID, raceID shared.RaceID, rosterMemberID *shared.MemberID) (RaceRecord, error)
}

// RaceSummaryRecord is a race as ListRaces reads it.
type RaceSummaryRecord struct {
	ID                    shared.RaceID
	Name                  string
	RaceDate              shared.Date
	Answer                race.Answer
	RunningOrderConfirmed bool // a running order of the race is confirmed
}

// RaceRecord is a race as FindRace reads it.
type RaceRecord struct {
	ID                    shared.RaceID
	Name                  string
	Version               int
	Description           race.DescriptionParams
	Answer                race.Answer
	RunningOrderConfirmed bool // a running order of the race is confirmed
	RunsInConfirmedOrder  bool // the roster member runs a leg of a confirmed running order
}
