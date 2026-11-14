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
	params DescriptionParams
}

// NewDescription validates the parts and returns the description.
func NewDescription(params DescriptionParams) (Description, error) {
	if params.RaceDate.IsZero() {
		return Description{}, shared.NewValidationError("race_date", "is required")
	}
	if params.Deadline.IsZero() {
		return Description{}, shared.NewValidationError("deadline", "is required")
	}
	switch params.Type {
	case Distance:
		if len(params.LegDistancesKm) == 0 {
			return Description{}, shared.NewValidationError("leg_distances_km", "needs at least one leg")
		}
		for _, distance := range params.LegDistancesKm {
			if distance <= 0 {
				return Description{}, shared.NewValidationError("leg_distances_km", "must be greater than 0")
			}
		}
		if params.TimeLimitMinutes != 0 || params.LapDistanceKm != 0 {
			return Description{}, shared.NewValidationError("type", "a distance race has no time limit or lap distance")
		}
		if params.TargetLaps != nil {
			return Description{}, shared.NewValidationError("target", "a distance race has a total time target")
		}
	case Time:
		if params.TimeLimitMinutes <= 0 {
			return Description{}, shared.NewValidationError("time_limit_minutes", "must be greater than 0")
		}
		if params.LapDistanceKm <= 0 {
			return Description{}, shared.NewValidationError("lap_distance_km", "must be greater than 0")
		}
		if len(params.LegDistancesKm) != 0 {
			return Description{}, shared.NewValidationError("type", "a time race has no leg distances")
		}
		if params.TargetTotalTime != nil {
			return Description{}, shared.NewValidationError("target", "a time race has a laps target")
		}
		if params.TargetLaps != nil && *params.TargetLaps <= 0 {
			return Description{}, shared.NewValidationError("target", "laps must be greater than 0")
		}
	default:
		return Description{}, shared.NewValidationError("type", "must be distance or time")
	}
	params.LegDistancesKm = slices.Clone(params.LegDistancesKm)
	return Description{params: params}, nil
}

// Type returns the race type.
func (description Description) Type() Type { return description.params.Type }

// RaceDate returns the race date.
func (description Description) RaceDate() shared.Date { return description.params.RaceDate }

// Deadline returns the deadline. Only organizers may see it (ADR-0006).
func (description Description) Deadline() time.Time { return description.params.Deadline }

// Text returns the free text.
func (description Description) Text() string { return description.params.Text }

// LegDistancesKm returns the distance of each leg (distance races).
func (description Description) LegDistancesKm() []float64 {
	return slices.Clone(description.params.LegDistancesKm)
}

// LegCount returns the number of legs of a distance race, and 0 for a time race.
func (description Description) LegCount() int { return len(description.params.LegDistancesKm) }

// TimeLimitMinutes returns the time limit (time races).
func (description Description) TimeLimitMinutes() int { return description.params.TimeLimitMinutes }

// LapDistanceKm returns the distance of one lap (time races).
func (description Description) LapDistanceKm() float64 { return description.params.LapDistanceKm }

// TargetTotalTime returns the target total time of a distance race, if set.
func (description Description) TargetTotalTime() (shared.TotalTime, bool) {
	if description.params.TargetTotalTime == nil {
		return shared.TotalTime{}, false
	}
	return *description.params.TargetTotalTime, true
}

// TargetLaps returns the target laps of a time race, if set.
func (description Description) TargetLaps() (int, bool) {
	if description.params.TargetLaps == nil {
		return 0, false
	}
	return *description.params.TargetLaps, true
}
