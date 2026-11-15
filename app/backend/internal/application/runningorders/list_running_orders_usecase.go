package runningorders

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
)

// ListRunningOrdersUseCase returns the running orders of a race as the caller may see them
// (GET /races/{race_id}/running-orders).
type ListRunningOrdersUseCase struct {
	query           RunningOrderQueryService
	raceVisibility  race.VisibilityPolicy
	orderVisibility runningorder.VisibilityPolicy
	predictor       Predictor
	dtoFactory      RunningOrderDTOFactory
	teamID          shared.TeamID
}

// NewListRunningOrdersUseCase returns a ListRunningOrdersUseCase for the team.
func NewListRunningOrdersUseCase(query RunningOrderQueryService, raceVisibility race.VisibilityPolicy, orderVisibility runningorder.VisibilityPolicy, predictor Predictor, dtoFactory RunningOrderDTOFactory, teamID shared.TeamID) *ListRunningOrdersUseCase {
	return &ListRunningOrdersUseCase{query: query, raceVisibility: raceVisibility, orderVisibility: orderVisibility, predictor: predictor, dtoFactory: dtoFactory, teamID: teamID}
}

// Execute returns the running orders with their predictions. Organizers see drafts and
// confirmed orders; members who answered "attending" see confirmed orders only, and other
// members of the race an empty list. A race the caller may not see is shared.ErrNotFound.
func (useCase *ListRunningOrdersUseCase) Execute(ctx context.Context, caller auth.MeDTO, raceID shared.RaceID) ([]RunningOrderDTO, error) {
	record, err := useCase.query.ListRunningOrders(ctx, useCase.teamID, raceID, rosterMemberOf(caller))
	if err != nil {
		return nil, err
	}
	if !useCase.raceVisibility.For(race.Viewer{Organizer: caller.IsOrganizer(), Answer: record.Answer}).Race {
		return nil, fmt.Errorf("race %s: %w", raceID, shared.ErrNotFound)
	}
	description, err := race.NewDescription(record.Description)
	if err != nil {
		return nil, fmt.Errorf("race %s has an invalid stored description: %w", raceID, err)
	}
	viewer := runningorder.Viewer{Organizer: caller.IsOrganizer(), Attending: record.Answer == race.Attending}
	runningOrders := []RunningOrderDTO{}
	for _, order := range record.RunningOrders {
		if !useCase.orderVisibility.CanSee(viewer, order.Status) {
			continue
		}
		prediction, err := useCase.predictor.Predict(description, order.Legs)
		if err != nil {
			return nil, err
		}
		runningOrders = append(runningOrders, useCase.dtoFactory.Create(order, description, prediction, caller.IsOrganizer()))
	}
	return runningOrders, nil
}
