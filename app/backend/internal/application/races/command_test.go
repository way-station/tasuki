package races

import (
	"errors"
	"testing"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func distanceInput() DescriptionInput {
	return DescriptionInput{
		Type:           "distance",
		RaceDate:       "2026-11-23",
		Deadline:       time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC),
		LegDistancesKm: []float64{5, 3.2},
	}
}

func timeInput() DescriptionInput {
	return DescriptionInput{
		Type:             "time",
		RaceDate:         "2026-11-23",
		Deadline:         time.Date(2026, 11, 1, 9, 0, 0, 0, time.UTC),
		TimeLimitMinutes: 180,
		LapDistanceKm:    2,
	}
}

func isValidation(err error) bool {
	var validationErr *shared.ValidationError
	return errors.As(err, &validationErr)
}

func TestRaceCommandFactory(t *testing.T) {
	t.Run("creates a race with a parsed description", func(t *testing.T) {
		input := distanceInput()
		totalTime := "2:45:00"
		input.TargetTotalTime = &totalTime
		command, err := RaceCommandFactory{}.CreateRace("Ekiden", input)
		if err != nil {
			t.Fatalf("CreateRace() error = %v", err)
		}
		if command.name != "Ekiden" || command.description.Type() != race.Distance || command.description.RaceDate().String() != "2026-11-23" {
			t.Errorf("CreateRace() = %+v", command)
		}
		if target, ok := command.description.TargetTotalTime(); !ok || target.String() != "2:45:00" {
			t.Errorf("TargetTotalTime() = %v, %v, want 2:45:00", target, ok)
		}
	})

	t.Run("updates a race with its ID and version", func(t *testing.T) {
		command, err := RaceCommandFactory{}.UpdateRace("r1", "Ekiden", timeInput(), 3)
		if err != nil {
			t.Fatalf("UpdateRace() error = %v", err)
		}
		if command.raceID != "r1" || command.version != 3 || command.description.Type() != race.Time {
			t.Errorf("UpdateRace() = %+v", command)
		}
	})

	tests := []struct {
		name string
		edit func(*DescriptionInput)
	}{
		{name: "malformed race date", edit: func(input *DescriptionInput) { input.RaceDate = "2026/11/23" }},
		{name: "malformed target time", edit: func(input *DescriptionInput) {
			totalTime := "165"
			input.TargetTotalTime = &totalTime
		}},
		{name: "unknown type", edit: func(input *DescriptionInput) { input.Type = "relay" }},
		{name: "no legs in a distance race", edit: func(input *DescriptionInput) { input.LegDistancesKm = nil }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := distanceInput()
			testCase.edit(&input)
			if _, err := (RaceCommandFactory{}).CreateRace("Ekiden", input); !isValidation(err) {
				t.Errorf("CreateRace() error = %v, want a validation error", err)
			}
			if _, err := (RaceCommandFactory{}).UpdateRace("r1", "Ekiden", input, 1); !isValidation(err) {
				t.Errorf("UpdateRace() error = %v, want a validation error", err)
			}
		})
	}
}
