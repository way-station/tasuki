//go:build integration

package postgres

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

func saveTestRace(t *testing.T, client *Client, teamID shared.TeamID, name string) shared.RaceID {
	t.Helper()
	factory := race.NewFactory(random.IDGenerator{})
	created, err := factory.New(teamID, name, distanceDescription(t, 5, 3.2))
	if err != nil {
		t.Fatal(err)
	}
	if err := NewRaceRepository(client, factory).Save(context.Background(), created); err != nil {
		t.Fatal(err)
	}
	return created.ID()
}

func saveTestRosterMember(t *testing.T, client *Client, name string) shared.MemberID {
	t.Helper()
	id := shared.MemberID(random.IDGenerator{}.NewID())
	mustExec(t, client, "INSERT INTO roster_members (id, team_id, name) VALUES ($1, $2, $3)", string(id), string(testTeamID), name)
	return id
}

func inviteTestMember(t *testing.T, client *Client, raceID shared.RaceID, memberID shared.MemberID, status string) {
	t.Helper()
	mustExec(t, client, "INSERT INTO race_invitations (race_id, roster_member_id) VALUES ($1, $2)", string(raceID), string(memberID))
	if status != "" {
		mustExec(t, client, "INSERT INTO attendances (race_id, roster_member_id, status) VALUES ($1, $2, $3)", string(raceID), string(memberID), status)
	}
}

// confirmTestRunningOrder creates an entry of members and a confirmed running order in
// which runner runs the only leg.
func confirmTestRunningOrder(t *testing.T, client *Client, raceID shared.RaceID, runner shared.MemberID, members ...shared.MemberID) {
	t.Helper()
	ids := random.IDGenerator{}
	entryID, runningOrderID := ids.NewID(), ids.NewID()
	mustExec(t, client, "INSERT INTO entries (id, race_id, name) VALUES ($1, $2, 'Team A')", entryID, string(raceID))
	for _, member := range members {
		mustExec(t, client, "INSERT INTO entry_members (entry_id, race_id, roster_member_id) VALUES ($1, $2, $3)", entryID, string(raceID), string(member))
	}
	mustExec(t, client, "INSERT INTO running_orders (id, entry_id, status, confirmed_at) VALUES ($1, $2, 'confirmed', now())", runningOrderID, entryID)
	mustExec(t, client, "INSERT INTO running_order_legs (running_order_id, entry_id, leg_number, roster_member_id) VALUES ($1, $2, 1, $3)", runningOrderID, entryID, string(runner))
}

func TestRaceQueryService(t *testing.T) {
	client := newTestClient(t)
	saveTestTeam(t, client)
	const otherTeamID shared.TeamID = "01900000-0000-7000-8000-000000000002"
	mustExec(t, client, "INSERT INTO teams (id, name) VALUES ($1, 'Other team')", string(otherTeamID))
	ctx := context.Background()
	queryService := NewRaceQueryService(client)

	runner := saveTestRosterMember(t, client, "Runner")
	teammate := saveTestRosterMember(t, client, "Teammate")
	notAnswered := saveTestRace(t, client, testTeamID, "Not answered")
	attending := saveTestRace(t, client, testTeamID, "Attending")
	notInvited := saveTestRace(t, client, testTeamID, "Not invited")
	notAttending := saveTestRace(t, client, testTeamID, "Not attending")
	otherTeams := saveTestRace(t, client, otherTeamID, "Other team's race")
	inviteTestMember(t, client, notAnswered, runner, "")
	inviteTestMember(t, client, attending, runner, "attending")
	inviteTestMember(t, client, attending, teammate, "attending")
	inviteTestMember(t, client, notAttending, runner, "not_attending")
	confirmTestRunningOrder(t, client, attending, runner, runner, teammate)

	t.Run("lists the team's races newest first with the roster member's answers", func(t *testing.T) {
		type summary struct {
			id        shared.RaceID
			answer    race.Answer
			confirmed bool
		}
		tests := []struct {
			name         string
			rosterMember *shared.MemberID
			want         []summary
		}{
			{name: "a roster member", rosterMember: &runner, want: []summary{
				{id: notAttending, answer: race.NotAttending},
				{id: notInvited, answer: race.NotInvited},
				{id: attending, answer: race.Attending, confirmed: true},
				{id: notAnswered, answer: race.NotAnswered},
			}},
			{name: "no roster member", rosterMember: nil, want: []summary{
				{id: notAttending, answer: race.NotInvited},
				{id: notInvited, answer: race.NotInvited},
				{id: attending, answer: race.NotInvited, confirmed: true},
				{id: notAnswered, answer: race.NotInvited},
			}},
		}
		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				records, err := queryService.ListRaces(ctx, testTeamID, testCase.rosterMember)
				if err != nil {
					t.Fatalf("ListRaces: %v", err)
				}
				got := make([]summary, 0, len(records))
				for _, record := range records {
					if record.RaceDate.String() != "2026-11-23" {
						t.Errorf("race date of %s = %s, want 2026-11-23", record.Name, record.RaceDate)
					}
					got = append(got, summary{id: record.ID, answer: record.Answer, confirmed: record.RunningOrderConfirmed})
				}
				if !slices.Equal(got, testCase.want) {
					t.Errorf("ListRaces = %+v, want %+v", got, testCase.want)
				}
			})
		}
	})

	t.Run("finds a race with whether the roster member runs in its confirmed running order", func(t *testing.T) {
		tests := []struct {
			name         string
			rosterMember *shared.MemberID
			answer       race.Answer
			runs         bool
		}{
			{name: "the runner", rosterMember: &runner, answer: race.Attending, runs: true},
			{name: "a teammate without a leg", rosterMember: &teammate, answer: race.Attending, runs: false},
			{name: "no roster member", rosterMember: nil, answer: race.NotInvited, runs: false},
		}
		for _, testCase := range tests {
			t.Run(testCase.name, func(t *testing.T) {
				record, err := queryService.FindRace(ctx, testTeamID, attending, testCase.rosterMember)
				if err != nil {
					t.Fatalf("FindRace: %v", err)
				}
				if record.ID != attending || record.Name != "Attending" || record.Version != 0 || record.Answer != testCase.answer ||
					!record.RunningOrderConfirmed || record.RunsInConfirmedOrder != testCase.runs {
					t.Errorf("FindRace = %+v, want answer %v and runs %v", record, testCase.answer, testCase.runs)
				}
				description := record.Description
				if description.Type != race.Distance || !slices.Equal(description.LegDistancesKm, []float64{5, 3.2}) || !description.Deadline.Equal(testNow) {
					t.Errorf("FindRace description = %+v", description)
				}
			})
		}
	})

	t.Run("does not find a race of another team", func(t *testing.T) {
		if _, err := queryService.FindRace(ctx, testTeamID, otherTeams, &runner); !errors.Is(err, shared.ErrNotFound) {
			t.Errorf("FindRace error = %v, want ErrNotFound", err)
		}
	})
}
