package races

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func organizer() auth.MeDTO {
	return auth.MeDTOFactory{}.Create("u1", "Organizer", true, identity.CallerNone, nil)
}

func member() auth.MeDTO {
	factory := auth.MeDTOFactory{}
	return factory.Create("u2", "Member", false, identity.CallerApproved, factory.CreateRosterMember("m1", "Member"))
}

type fakeRaceQueryService struct {
	summaries       []RaceSummaryRecord
	record          RaceRecord
	err             error
	gotTeamID       shared.TeamID
	gotRosterMember *shared.MemberID
}

func (fake *fakeRaceQueryService) ListRaces(_ context.Context, teamID shared.TeamID, rosterMemberID *shared.MemberID) ([]RaceSummaryRecord, error) {
	fake.gotTeamID, fake.gotRosterMember = teamID, rosterMemberID
	return fake.summaries, fake.err
}

func (fake *fakeRaceQueryService) FindRace(_ context.Context, teamID shared.TeamID, raceID shared.RaceID, rosterMemberID *shared.MemberID) (RaceRecord, error) {
	fake.gotTeamID, fake.gotRosterMember = teamID, rosterMemberID
	if fake.err != nil {
		return RaceRecord{}, fake.err
	}
	if fake.record.ID != raceID {
		return RaceRecord{}, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	return fake.record, nil
}

func summaryRecords() []RaceSummaryRecord {
	answers := []race.Answer{race.NotInvited, race.NotAnswered, race.Undecided, race.Attending, race.NotAttending}
	records := make([]RaceSummaryRecord, 0, len(answers))
	for index, answer := range answers {
		records = append(records, RaceSummaryRecord{ID: shared.RaceID(fmt.Sprintf("r%d", index+1)), Name: "Race", Answer: answer, RunningOrderConfirmed: true})
	}
	return records
}

func newListRacesUseCase(query RaceQueryService) *ListRacesUseCase {
	return NewListRacesUseCase(query, race.VisibilityPolicy{}, RaceDTOFactory{}, teamID)
}

func TestListRacesUseCase(t *testing.T) {
	type want struct {
		id        shared.RaceID
		invited   bool
		confirmed bool // whether running_order_confirmed is returned
	}
	tests := []struct {
		name         string
		caller       auth.MeDTO
		rosterMember *shared.MemberID
		want         []want
	}{
		{
			name:   "organizers see every race",
			caller: organizer(),
			want: []want{
				{id: "r1", invited: false, confirmed: true},
				{id: "r2", invited: true, confirmed: true},
				{id: "r3", invited: true, confirmed: true},
				{id: "r4", invited: true, confirmed: true},
				{id: "r5", invited: true, confirmed: true},
			},
		},
		{
			name:         "members see invited races except not attending, and the running order when attending",
			caller:       member(),
			rosterMember: new(shared.MemberID("m1")),
			want: []want{
				{id: "r2", invited: true, confirmed: false},
				{id: "r3", invited: true, confirmed: false},
				{id: "r4", invited: true, confirmed: true},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query := &fakeRaceQueryService{summaries: summaryRecords()}
			summaries, err := newListRacesUseCase(query).Execute(context.Background(), testCase.caller)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if query.gotTeamID != teamID || !samePointer(query.gotRosterMember, testCase.rosterMember) {
				t.Errorf("queried %s for %v, want %s for %v", query.gotTeamID, query.gotRosterMember, teamID, testCase.rosterMember)
			}
			got := make([]want, 0, len(summaries))
			for _, summary := range summaries {
				_, invited := summary.MyAttendance()
				_, confirmed := summary.RunningOrderConfirmed()
				got = append(got, want{id: summary.ID(), invited: invited, confirmed: confirmed})
			}
			if !slices.Equal(got, testCase.want) {
				t.Errorf("Execute() = %+v, want %+v", got, testCase.want)
			}
		})
	}

	t.Run("returns the query error", func(t *testing.T) {
		queryErr := errors.New("database is down")
		if _, err := newListRacesUseCase(&fakeRaceQueryService{err: queryErr}).Execute(context.Background(), member()); !errors.Is(err, queryErr) {
			t.Errorf("Execute() error = %v, want %v", err, queryErr)
		}
	})
}

func samePointer(got, want *shared.MemberID) bool {
	if got == nil || want == nil {
		return got == want
	}
	return *got == *want
}
