package race

import (
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type fixedIDs struct{}

func (fixedIDs) NewID() string { return "r1" }

func distanceParams(t *testing.T) DescriptionParams {
	t.Helper()
	date, err := shared.ParseDate("2026-11-23")
	if err != nil {
		t.Fatal(err)
	}
	return DescriptionParams{
		Type:           Distance,
		RaceDate:       date,
		Deadline:       time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC),
		LegDistancesKm: []float64{5, 3.2},
	}
}

func isValidation(err error) bool {
	var validationErr *shared.ValidationError
	return errors.As(err, &validationErr)
}

func TestNewDescription(t *testing.T) {
	laps := 30
	total, _ := shared.ParseTotalTime("2:45:00")

	tests := []struct {
		name    string
		edit    func(*DescriptionParams)
		wantErr bool
	}{
		{name: "distance", edit: func(*DescriptionParams) {}},
		{name: "distance with total time target", edit: func(params *DescriptionParams) { params.TargetTotalTime = &total }},
		{name: "distance without legs", edit: func(params *DescriptionParams) { params.LegDistancesKm = nil }, wantErr: true},
		{name: "distance with a zero leg", edit: func(params *DescriptionParams) { params.LegDistancesKm = []float64{5, 0} }, wantErr: true},
		{name: "distance with laps target", edit: func(params *DescriptionParams) { params.TargetLaps = &laps }, wantErr: true},
		{name: "distance with time limit", edit: func(params *DescriptionParams) { params.TimeLimitMinutes = 180 }, wantErr: true},
		{name: "time", edit: func(params *DescriptionParams) {
			params.Type, params.LegDistancesKm, params.TimeLimitMinutes, params.LapDistanceKm, params.TargetLaps = Time, nil, 180, 2, &laps
		}},
		{name: "time without lap distance", edit: func(params *DescriptionParams) {
			params.Type, params.LegDistancesKm, params.TimeLimitMinutes = Time, nil, 180
		}, wantErr: true},
		{name: "time with legs", edit: func(params *DescriptionParams) {
			params.Type, params.TimeLimitMinutes, params.LapDistanceKm = Time, 180, 2
		}, wantErr: true},
		{name: "no deadline", edit: func(params *DescriptionParams) { params.Deadline = time.Time{} }, wantErr: true},
		{name: "unknown type", edit: func(params *DescriptionParams) { params.Type = "relay" }, wantErr: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			params := distanceParams(t)
			testCase.edit(&params)
			_, err := NewDescription(params)
			if testCase.wantErr != isValidation(err) {
				t.Errorf("NewDescription() error = %v, want a ValidationError: %v", err, testCase.wantErr)
			}
		})
	}
}

func TestDescriptionCopiesLegs(t *testing.T) {
	params := distanceParams(t)
	description, err := NewDescription(params)
	if err != nil {
		t.Fatal(err)
	}
	params.LegDistancesKm[0] = 99
	description.LegDistancesKm()[1] = 99
	if got := description.LegDistancesKm(); got[0] != 5 || got[1] != 3.2 || description.LegCount() != 2 {
		t.Errorf("LegDistancesKm() = %v; the description must not share its slice", got)
	}
}

func TestRaceUpdate(t *testing.T) {
	description, err := NewDescription(distanceParams(t))
	if err != nil {
		t.Fatal(err)
	}
	race, err := NewFactory(fixedIDs{}).New("team1", "Ekiden", description)
	if err != nil {
		t.Fatal(err)
	}
	if err := race.Update("Ekiden 2026", description, 1); !errors.Is(err, shared.ErrVersionConflict) {
		t.Errorf("Update(stale version) = %v, want ErrVersionConflict", err)
	}
	if err := race.Update(" ", description, 0); !isValidation(err) {
		t.Errorf("Update(blank name) = %v, want a ValidationError", err)
	}
	if err := race.Update("Ekiden 2026", description, 0); err != nil || race.Name() != "Ekiden 2026" || race.Version() != 1 {
		t.Errorf("Update() = %v, name %q, version %description", err, race.Name(), race.Version())
	}
}

func TestVisibilityPolicy(t *testing.T) {
	tests := []struct {
		name   string
		viewer Viewer
		want   Visible
	}{
		{name: "organizer", viewer: Viewer{Organizer: true}, want: Visible{Race: true, Description: true, Deadline: true, RunningOrderSummary: true}},
		{name: "not invited", viewer: Viewer{Answer: NotInvited}, want: Visible{}},
		{name: "not answered", viewer: Viewer{Answer: NotAnswered}, want: Visible{Race: true, Description: true}},
		{name: "undecided", viewer: Viewer{Answer: Undecided}, want: Visible{Race: true, Description: true}},
		{name: "attending", viewer: Viewer{Answer: Attending}, want: Visible{Race: true, Description: true, RunningOrderSummary: true}},
		{name: "not attending", viewer: Viewer{Answer: NotAttending}, want: Visible{}},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			if got := (VisibilityPolicy{}).For(testCase.viewer); got != testCase.want {
				t.Errorf("For() = %+v, want %+v", got, testCase.want)
			}
		})
	}
}
