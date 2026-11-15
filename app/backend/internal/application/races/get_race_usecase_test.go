package races

import (
	"context"
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func raceRecord(t *testing.T, answer race.Answer) RaceRecord {
	t.Helper()
	date, err := shared.ParseDate("2026-11-23")
	if err != nil {
		t.Fatal(err)
	}
	input := distanceInput()
	return RaceRecord{
		ID:      "r1",
		Name:    "Ekiden",
		Version: 3,
		Description: race.DescriptionParams{
			Type: race.Distance, RaceDate: date, Deadline: input.Deadline, LegDistancesKm: input.LegDistancesKm,
		},
		Answer:                answer,
		RunningOrderConfirmed: true,
		RunsInConfirmedOrder:  true,
	}
}

func newGetRaceUseCase(query RaceQueryService) *GetRaceUseCase {
	return NewGetRaceUseCase(query, race.VisibilityPolicy{}, RaceDTOFactory{}, teamID)
}

func TestGetRaceUseCase(t *testing.T) {
	tests := []struct {
		name          string
		caller        auth.MeDTO
		answer        race.Answer
		wantVersion   bool
		wantDeadline  bool
		wantConfirmed bool
		wantInvited   bool
	}{
		{name: "organizers see everything", caller: organizer(), answer: race.NotInvited, wantVersion: true, wantDeadline: true, wantConfirmed: true, wantInvited: false},
		{name: "attending members see the running order", caller: member(), answer: race.Attending, wantConfirmed: true, wantInvited: true},
		{name: "members who have not answered see the race only", caller: member(), answer: race.NotAnswered, wantInvited: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			query := &fakeRaceQueryService{record: raceRecord(t, testCase.answer)}
			got, err := newGetRaceUseCase(query).Execute(context.Background(), testCase.caller, "r1")
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}
			if query.gotTeamID != teamID || !samePointer(query.gotRosterMember, rosterMemberOf(testCase.caller)) {
				t.Errorf("queried %s for %v, want %s for the caller's roster member", query.gotTeamID, query.gotRosterMember, teamID)
			}
			if got.ID() != "r1" || got.Name() != "Ekiden" || got.Description().Type() != race.Distance || !got.CanSubmitLaps() {
				t.Errorf("Execute() = %+v", got)
			}
			if version, ok := got.Version(); ok != testCase.wantVersion || (ok && version != 3) {
				t.Errorf("Version() = %d, %v, want shown %v", version, ok, testCase.wantVersion)
			}
			if _, ok := got.Description().Deadline(); ok != testCase.wantDeadline {
				t.Errorf("Deadline shown = %v, want %v", ok, testCase.wantDeadline)
			}
			if confirmed, ok := got.RunningOrderConfirmed(); ok != testCase.wantConfirmed || (ok && !confirmed) {
				t.Errorf("RunningOrderConfirmed() = %v, %v, want shown %v", confirmed, ok, testCase.wantConfirmed)
			}
			if answer, ok := got.MyAttendance(); ok != testCase.wantInvited || answer != testCase.answer {
				t.Errorf("MyAttendance() = %v, %v, want %v, %v", answer, ok, testCase.answer, testCase.wantInvited)
			}
		})
	}

	hidden := []struct {
		name   string
		answer race.Answer
		raceID shared.RaceID
	}{
		{name: "members who are not invited", answer: race.NotInvited, raceID: "r1"},
		{name: "members who answered not attending", answer: race.NotAttending, raceID: "r1"},
		{name: "a race that does not exist", answer: race.Attending, raceID: "r9"},
	}
	for _, testCase := range hidden {
		t.Run("not found for "+testCase.name, func(t *testing.T) {
			query := &fakeRaceQueryService{record: raceRecord(t, testCase.answer)}
			if _, err := newGetRaceUseCase(query).Execute(context.Background(), member(), testCase.raceID); !errors.Is(err, shared.ErrNotFound) {
				t.Errorf("Execute() error = %v, want shared.ErrNotFound", err)
			}
		})
	}

	t.Run("fails on an invalid stored description", func(t *testing.T) {
		record := raceRecord(t, race.Attending)
		record.Description.LegDistancesKm = nil
		_, err := newGetRaceUseCase(&fakeRaceQueryService{record: record}).Execute(context.Background(), member(), "r1")
		if err == nil || errors.Is(err, shared.ErrNotFound) {
			t.Errorf("Execute() error = %v, want an error that is not shared.ErrNotFound", err)
		}
	})
}
