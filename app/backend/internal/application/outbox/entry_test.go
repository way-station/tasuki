package outbox

import (
	"errors"
	"testing"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

type namedStep string

func (step namedStep) Kind() string { return string(step) }

func TestEntryFactory(t *testing.T) {
	t.Run("keeps the steps in order", func(t *testing.T) {
		entry, err := EntryFactory{}.New(namedStep("first"), namedStep("second"))
		if err != nil {
			t.Fatalf("New() error = %v", err)
		}
		steps := entry.Steps()
		if len(steps) != 2 || steps[0].Kind() != "first" || steps[1].Kind() != "second" {
			t.Errorf("Steps() = %v, want first and second", steps)
		}
	})

	tests := []struct {
		name  string
		steps []Step
	}{
		{name: "no step", steps: nil},
		{name: "a step without a kind", steps: []Step{namedStep("first"), namedStep("")}},
		{name: "a nil step", steps: []Step{nil}},
	}
	for _, testCase := range tests {
		t.Run("rejects "+testCase.name, func(t *testing.T) {
			var validationErr *shared.ValidationError
			if _, err := (EntryFactory{}).New(testCase.steps...); !errors.As(err, &validationErr) {
				t.Errorf("New() error = %v, want a validation error", err)
			}
		})
	}
}
