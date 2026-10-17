// Package race is the domain of races: the race, its description, and who can see what
// (docs/design/database.md, ADR-0006).
package race

import (
	"slices"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Type is whether a race is distance-based or time-based.
type Type string

// Race types.
const (
	Distance Type = "distance"
	Time     Type = "time"
)

// DescriptionParams are the parts of a race description, as given by an organizer.
// Fields that do not belong to the race type must be left empty.
type DescriptionParams struct {
	Type             Type
	RaceDate         shared.Date
	Deadline         time.Time
	Text             string
	LegDistancesKm   []float64         // distance races
	TimeLimitMinutes int               // time races
	LapDistanceKm    float64           // time races
	TargetTotalTime  *shared.TotalTime // distance races, optional
	TargetLaps       *int              // time races, optional
}

// Description is a race description: a value object that always matches its race type.
type Description struct {
	p DescriptionParams
}

// NewDescription validates the parts and returns the description.
func NewDescription(p DescriptionParams) (Description, error) {
	if p.RaceDate.IsZero() {
		return Description{}, shared.NewValidationError("race_date", "is required")
	}
	if p.Deadline.IsZero() {
		return Description{}, shared.NewValidationError("deadline", "is required")
	}
	switch p.Type {
	case Distance:
		if len(p.LegDistancesKm) == 0 {
			return Description{}, shared.NewValidationError("leg_distances_km", "needs at least one leg")
		}
		for _, d := range p.LegDistancesKm {
			if d <= 0 {
				return Description{}, shared.NewValidationError("leg_distances_km", "must be greater than 0")
			}
		}
		if p.TimeLimitMinutes != 0 || p.LapDistanceKm != 0 {
			return Description{}, shared.NewValidationError("type", "a distance race has no time limit or lap distance")
		}
		if p.TargetLaps != nil {
			return Description{}, shared.NewValidationError("target", "a distance race has a total time target")
		}
	case Time:
		if p.TimeLimitMinutes <= 0 {
			return Description{}, shared.NewValidationError("time_limit_minutes", "must be greater than 0")
		}
		if p.LapDistanceKm <= 0 {
			return Description{}, shared.NewValidationError("lap_distance_km", "must be greater than 0")
		}
		if len(p.LegDistancesKm) != 0 {
			return Description{}, shared.NewValidationError("type", "a time race has no leg distances")
		}
		if p.TargetTotalTime != nil {
			return Description{}, shared.NewValidationError("target", "a time race has a laps target")
		}
		if p.TargetLaps != nil && *p.TargetLaps <= 0 {
			return Description{}, shared.NewValidationError("target", "laps must be greater than 0")
		}
	default:
		return Description{}, shared.NewValidationError("type", "must be distance or time")
	}
	p.LegDistancesKm = slices.Clone(p.LegDistancesKm)
	return Description{p: p}, nil
}

// Type returns the race type.
func (d Description) Type() Type { return d.p.Type }

// RaceDate returns the race date.
func (d Description) RaceDate() shared.Date { return d.p.RaceDate }

// Deadline returns the deadline. Only organizers may see it (ADR-0006).
func (d Description) Deadline() time.Time { return d.p.Deadline }

// Text returns the free text.
func (d Description) Text() string { return d.p.Text }

// LegDistancesKm returns the distance of each leg (distance races).
func (d Description) LegDistancesKm() []float64 { return slices.Clone(d.p.LegDistancesKm) }

// LegCount returns the number of legs of a distance race, and 0 for a time race.
func (d Description) LegCount() int { return len(d.p.LegDistancesKm) }

// TimeLimitMinutes returns the time limit (time races).
func (d Description) TimeLimitMinutes() int { return d.p.TimeLimitMinutes }

// LapDistanceKm returns the distance of one lap (time races).
func (d Description) LapDistanceKm() float64 { return d.p.LapDistanceKm }

// TargetTotalTime returns the target total time of a distance race, if set.
func (d Description) TargetTotalTime() (shared.TotalTime, bool) {
	if d.p.TargetTotalTime == nil {
		return shared.TotalTime{}, false
	}
	return *d.p.TargetTotalTime, true
}

// TargetLaps returns the target laps of a time race, if set.
func (d Description) TargetLaps() (int, bool) {
	if d.p.TargetLaps == nil {
		return 0, false
	}
	return *d.p.TargetLaps, true
}
