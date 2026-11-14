package outbox

import (
	"slices"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// Entry is the remaining work of one use case: its steps in the order they run. It is
// created by EntryFactory.
type Entry struct {
	steps []Step
}

// Steps returns the steps in the order they run.
func (entry Entry) Steps() []Step { return slices.Clone(entry.steps) }

// EntryFactory assembles Entry values.
type EntryFactory struct{}

// New returns the entry of the steps. An entry needs at least one step, and every step a
// kind.
func (EntryFactory) New(steps ...Step) (Entry, error) {
	if len(steps) == 0 {
		return Entry{}, shared.NewValidationError("steps", "needs at least one step")
	}
	for _, step := range steps {
		if step == nil || step.Kind() == "" {
			return Entry{}, shared.NewValidationError("steps", "every step needs a kind")
		}
	}
	return Entry{steps: slices.Clone(steps)}, nil
}
