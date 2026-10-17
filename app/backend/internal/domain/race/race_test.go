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
	var ve *shared.ValidationError
	return errors.As(err, &ve)
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
		{name: "distance with total time target", edit: func(p *DescriptionParams) { p.TargetTotalTime = &total }},
		{name: "distance without legs", edit: func(p *DescriptionParams) { p.LegDistancesKm = nil }, wantErr: true},
		{name: "distance with a zero leg", edit: func(p *DescriptionParams) { p.LegDistancesKm = []float64{5, 0} }, wantErr: true},
		{name: "distance with laps target", edit: func(p *DescriptionParams) { p.TargetLaps = &laps }, wantErr: true},
		{name: "distance with time limit", edit: func(p *DescriptionParams) { p.TimeLimitMinutes = 180 }, wantErr: true},
		{name: "time", edit: func(p *DescriptionParams) {
			p.Type, p.LegDistancesKm, p.TimeLimitMinutes, p.LapDistanceKm, p.TargetLaps = Time, nil, 180, 2, &laps
		}},
		{name: "time without lap distance", edit: func(p *DescriptionParams) {
			p.Type, p.LegDistancesKm, p.TimeLimitMinutes = Time, nil, 180
		}, wantErr: true},
		{name: "time with legs", edit: func(p *DescriptionParams) {
			p.Type, p.TimeLimitMinutes, p.LapDistanceKm = Time, 180, 2
		}, wantErr: true},
		{name: "no deadline", edit: func(p *DescriptionParams) { p.Deadline = time.Time{} }, wantErr: true},
		{name: "unknown type", edit: func(p *DescriptionParams) { p.Type = "relay" }, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := distanceParams(t)
			tt.edit(&p)
			_, err := NewDescription(p)
			if tt.wantErr != isValidation(err) {
				t.Errorf("NewDescription() error = %v, want a ValidationError: %v", err, tt.wantErr)
			}
		})
	}
}

func TestDescriptionCopiesLegs(t *testing.T) {
	p := distanceParams(t)
	d, err := NewDescription(p)
	if err != nil {
		t.Fatal(err)
	}
	p.LegDistancesKm[0] = 99
	d.LegDistancesKm()[1] = 99
	if got := d.LegDistancesKm(); got[0] != 5 || got[1] != 3.2 || d.LegCount() != 2 {
		t.Errorf("LegDistancesKm() = %v; the description must not share its slice", got)
	}
}

func TestRaceUpdate(t *testing.T) {
	d, err := NewDescription(distanceParams(t))
	if err != nil {
		t.Fatal(err)
	}
	r, err := NewFactory(fixedIDs{}).New("team1", "Ekiden", d)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Update("Ekiden 2026", d, 1); !errors.Is(err, shared.ErrVersionConflict) {
		t.Errorf("Update(stale version) = %v, want ErrVersionConflict", err)
	}
	if err := r.Update(" ", d, 0); !isValidation(err) {
		t.Errorf("Update(blank name) = %v, want a ValidationError", err)
	}
	if err := r.Update("Ekiden 2026", d, 0); err != nil || r.Name() != "Ekiden 2026" || r.Version() != 1 {
		t.Errorf("Update() = %v, name %q, version %d", err, r.Name(), r.Version())
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (VisibilityPolicy{}).For(tt.viewer); got != tt.want {
				t.Errorf("For() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
