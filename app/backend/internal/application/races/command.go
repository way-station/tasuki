// Package races holds the use cases of races: registering, editing, listing, and showing
// them (docs/design/api.md, ADR-0006, ADR-0011).
package races

import (
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// DescriptionInput is a race description as the request sends it. Fields that do not
// belong to the race type are left empty.
type DescriptionInput struct {
	Type             string
	RaceDate         string // YYYY-MM-DD
	Deadline         time.Time
	Text             string
	LegDistancesKm   []float64
	TimeLimitMinutes int
	LapDistanceKm    float64
	TargetTotalTime  *string // H:MM:SS
	TargetLaps       *int
}

// CreateRaceCommand is a race to register. It is created by RaceCommandFactory.
type CreateRaceCommand struct {
	name        string
	description race.Description
}

// UpdateRaceCommand is a race edit: the new name and description, and the version the
// organizer read. It is created by RaceCommandFactory.
type UpdateRaceCommand struct {
	raceID      shared.RaceID
	name        string
	description race.Description
	version     int
}

// RaceCommandFactory assembles CreateRaceCommand and UpdateRaceCommand values.
type RaceCommandFactory struct{}

// CreateRace returns the command, or a shared.ValidationError when the description is
// malformed or does not match its type.
func (RaceCommandFactory) CreateRace(name string, input DescriptionInput) (CreateRaceCommand, error) {
	description, err := newDescription(input)
	if err != nil {
		return CreateRaceCommand{}, err
	}
	return CreateRaceCommand{name: name, description: description}, nil
}

// UpdateRace returns the command, or a shared.ValidationError when the description is
// malformed or does not match its type.
func (RaceCommandFactory) UpdateRace(raceID shared.RaceID, name string, input DescriptionInput, version int) (UpdateRaceCommand, error) {
	description, err := newDescription(input)
	if err != nil {
		return UpdateRaceCommand{}, err
	}
	return UpdateRaceCommand{raceID: raceID, name: name, description: description, version: version}, nil
}

// newDescription parses the formats and leaves the rules of the description to the domain.
// The name is checked by the domain too, when the race is created or updated.
func newDescription(input DescriptionInput) (race.Description, error) {
	raceDate, err := shared.ParseDate(input.RaceDate)
	if err != nil {
		return race.Description{}, shared.NewValidationError("race_date", "must be YYYY-MM-DD")
	}
	var targetTotalTime *shared.TotalTime
	if input.TargetTotalTime != nil {
		totalTime, err := shared.ParseTotalTime(*input.TargetTotalTime)
		if err != nil {
			return race.Description{}, shared.NewValidationError("target", "total_time must be H:MM:SS")
		}
		targetTotalTime = &totalTime
	}
	return race.NewDescription(race.DescriptionParams{
		Type:             race.Type(input.Type),
		RaceDate:         raceDate,
		Deadline:         input.Deadline,
		Text:             input.Text,
		LegDistancesKm:   input.LegDistancesKm,
		TimeLimitMinutes: input.TimeLimitMinutes,
		LapDistanceKm:    input.LapDistanceKm,
		TargetTotalTime:  targetTotalTime,
		TargetLaps:       input.TargetLaps,
	})
}
