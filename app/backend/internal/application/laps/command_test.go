package laps

import (
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

func isValidation(err error) bool {
	var validationErr *shared.ValidationError
	return errors.As(err, &validationErr)
}

func TestLapCommandFactorySubmit(t *testing.T) {
	t.Run("parses the times in order", func(t *testing.T) {
		command, err := LapCommandFactory{}.Submit("r1", []string{"3:20", "3:25"})
		if err != nil {
			t.Fatalf("Submit() error = %v", err)
		}
		if command.raceID != "r1" || len(command.times) != 2 || command.times[0].String() != "3:20" || command.times[1].String() != "3:25" {
			t.Errorf("Submit() = %+v, want 3:20 and 3:25 for r1", command)
		}
	})

	tests := []struct {
		name  string
		times []string
		field string
	}{
		{name: "no time", times: nil, field: "times"},
		{name: "a time that is not MM:SS", times: []string{"3:20", "3:2"}, field: "times[1]"},
		{name: "a time of zero", times: []string{"0:00"}, field: "times[0]"},
	}
	for _, testCase := range tests {
		t.Run("rejects "+testCase.name, func(t *testing.T) {
			_, err := LapCommandFactory{}.Submit("r1", testCase.times)
			var validationErr *shared.ValidationError
			if !errors.As(err, &validationErr) || validationErr.Field != testCase.field {
				t.Errorf("Submit() error = %v, want a validation error of %s", err, testCase.field)
			}
		})
	}
}

func TestLapCommandFactoryEdit(t *testing.T) {
	t.Run("parses the time", func(t *testing.T) {
		command, err := LapCommandFactory{}.Edit("r1", "l1", "3:28")
		if err != nil {
			t.Fatalf("Edit() error = %v", err)
		}
		if command.raceID != "r1" || command.lapID != "l1" || command.lapTime.String() != "3:28" {
			t.Errorf("Edit() = %+v, want 3:28 for l1 of r1", command)
		}
	})

	t.Run("rejects a time that is not MM:SS", func(t *testing.T) {
		if _, err := (LapCommandFactory{}).Edit("r1", "l1", "328"); !isValidation(err) {
			t.Errorf("Edit() error = %v, want a validation error", err)
		}
	})
}

func TestLapCommandFactoryDelete(t *testing.T) {
	command := LapCommandFactory{}.Delete("r1", "l1")
	if command.raceID != "r1" || command.lapID != "l1" {
		t.Errorf("Delete() = %+v, want l1 of r1", command)
	}
}
