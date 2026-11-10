package runningorders

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// GetRunningOrderUseCase returns the running order of one entry as the caller may see it.
// The save and confirm endpoints return the running order through it.
type GetRunningOrderUseCase struct {
	list *ListRunningOrdersUseCase
}

// NewGetRunningOrderUseCase returns a GetRunningOrderUseCase that reads through list.
func NewGetRunningOrderUseCase(list *ListRunningOrdersUseCase) *GetRunningOrderUseCase {
	return &GetRunningOrderUseCase{list: list}
}

// Execute returns the running order of the entry. An entry without a running order the
// caller may see is shared.ErrNotFound.
func (useCase *GetRunningOrderUseCase) Execute(ctx context.Context, caller auth.MeDTO, raceID shared.RaceID, entryID shared.EntryID) (RunningOrderDTO, error) {
	runningOrders, err := useCase.list.Execute(ctx, caller, raceID)
	if err != nil {
		return RunningOrderDTO{}, err
	}
	for _, runningOrder := range runningOrders {
		if runningOrder.EntryID() == entryID {
			return runningOrder, nil
		}
	}
	return RunningOrderDTO{}, fmt.Errorf("running order of entry %s: %w", entryID, shared.ErrNotFound)
}
