package controller

import (
	"context"
	"errors"

	"github.com/oapi-codegen/nullable"

	"github.com/way-station/tasuki/app/backend/internal/application/runningorders"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

// RunningOrderController serves the running orders and their predictions
// (docs/design/api.md). The routes check the caller's role before it runs; reads apply
// what the caller may see in the use cases.
type RunningOrderController struct {
	list           *runningorders.ListRunningOrdersUseCase
	get            *runningorders.GetRunningOrderUseCase
	save           *runningorders.SaveRunningOrderUseCase
	confirm        *runningorders.ConfirmRunningOrderUseCase
	commandFactory runningorders.RunningOrderCommandFactory
}

// NewRunningOrderController returns a RunningOrderController.
func NewRunningOrderController(list *runningorders.ListRunningOrdersUseCase, get *runningorders.GetRunningOrderUseCase, save *runningorders.SaveRunningOrderUseCase, confirm *runningorders.ConfirmRunningOrderUseCase, commandFactory runningorders.RunningOrderCommandFactory) *RunningOrderController {
	return &RunningOrderController{list: list, get: get, save: save, confirm: confirm, commandFactory: commandFactory}
}

// ListRunningOrders returns the running orders of the race the caller may see. A race the
// caller may not see is 404.
func (controller *RunningOrderController) ListRunningOrders(ctx context.Context, request api.ListRunningOrdersRequestObject) (api.ListRunningOrdersResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.ListRunningOrders401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	runningOrders, err := controller.list.Execute(ctx, caller, shared.RaceID(request.RaceId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.ListRunningOrders404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	response := api.ListRunningOrders200JSONResponse{RunningOrders: make([]api.RunningOrder, 0, len(runningOrders))}
	for _, runningOrder := range runningOrders {
		response.RunningOrders = append(response.RunningOrders, toAPIRunningOrder(runningOrder))
	}
	return response, nil
}

// SaveRunningOrder saves the legs of the entry's running order and returns it.
func (controller *RunningOrderController) SaveRunningOrder(ctx context.Context, request api.SaveRunningOrderRequestObject) (api.SaveRunningOrderResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.SaveRunningOrder401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	if request.Body == nil {
		return api.SaveRunningOrder400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	raceID, entryID := shared.RaceID(request.RaceId), shared.EntryID(request.EntryId)
	err := controller.save.Execute(ctx, caller, controller.commandFactory.Save(raceID, entryID, request.Body.Version, toLegInputs(request.Body.Legs)))
	switch {
	case errors.Is(err, shared.ErrNotFound):
		return api.SaveRunningOrder404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	case errors.Is(err, shared.ErrVersionConflict):
		return api.SaveRunningOrder409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: versionConflict()}, nil
	}
	if problem, ok := validationFailed(err); ok {
		return api.SaveRunningOrder422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: problem}, nil
	}
	if err != nil {
		return nil, err
	}
	saved, err := controller.get.Execute(ctx, caller, raceID, entryID)
	if err != nil {
		return nil, err
	}
	return api.SaveRunningOrder200JSONResponse(toAPIRunningOrder(saved)), nil
}

// ConfirmRunningOrder confirms the entry's running order and returns it.
func (controller *RunningOrderController) ConfirmRunningOrder(ctx context.Context, request api.ConfirmRunningOrderRequestObject) (api.ConfirmRunningOrderResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.ConfirmRunningOrder401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	if request.Body == nil {
		return api.ConfirmRunningOrder400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	raceID, entryID := shared.RaceID(request.RaceId), shared.EntryID(request.EntryId)
	err := controller.confirm.Execute(ctx, caller, controller.commandFactory.Confirm(raceID, entryID, request.Body.Version))
	switch {
	case errors.Is(err, shared.ErrNotFound):
		return api.ConfirmRunningOrder404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	case errors.Is(err, shared.ErrVersionConflict):
		return api.ConfirmRunningOrder409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: versionConflict()}, nil
	}
	if problem, ok := stateConflict(err); ok {
		return api.ConfirmRunningOrder409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: problem}, nil
	}
	if problem, ok := validationFailed(err); ok {
		return api.ConfirmRunningOrder422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: problem}, nil
	}
	if err != nil {
		return nil, err
	}
	confirmed, err := controller.get.Execute(ctx, caller, raceID, entryID)
	if err != nil {
		return nil, err
	}
	return api.ConfirmRunningOrder200JSONResponse(toAPIRunningOrder(confirmed)), nil
}

// toLegInputs reads the legs of the request. A leg whose roster_member_id is null or left
// out is empty.
func toLegInputs(legs []api.LegAssignment) []runningorders.LegInput {
	inputs := make([]runningorders.LegInput, 0, len(legs))
	for _, leg := range legs {
		input := runningorders.LegInput{Number: leg.LegNumber}
		if runnerID, err := leg.RosterMemberId.Get(); err == nil {
			runner := shared.MemberID(runnerID)
			input.Runner = &runner
		}
		inputs = append(inputs, input)
	}
	return inputs
}

func toAPIRunningOrder(dto runningorders.RunningOrderDTO) api.RunningOrder {
	legs := dto.Legs()
	body := api.RunningOrder{EntryId: string(dto.EntryID()), EntryName: dto.EntryName(), Status: api.RunningOrderStatus(dto.Status()), Legs: make([]api.Leg, 0, len(legs))}
	if version, ok := dto.Version(); ok {
		body.Version = &version
	}
	for _, leg := range legs {
		apiLeg := api.Leg{LegNumber: leg.Number(), Runner: nullable.NewNullNullable[api.RosterMemberRef]()}
		if distance, ok := leg.DistanceKm(); ok {
			distanceKm := float32(distance)
			apiLeg.DistanceKm = &distanceKm
		}
		if runner, ok := leg.Runner(); ok {
			apiLeg.Runner = nullable.NewNullableWithValue(api.RosterMemberRef{RosterMemberId: string(runner.MemberID()), Name: runner.Name()})
		}
		body.Legs = append(body.Legs, apiLeg)
	}
	body.Prediction = nullable.NewNullNullable[api.Prediction]()
	if prediction, ok := dto.Prediction(); ok {
		apiPrediction := api.Prediction{HasEmptyLegs: prediction.HasEmptyLegs(), TotalTime: nullable.NewNullNullable[api.HMmSs](), Laps: nullable.NewNullNullable[int]()}
		if totalTime, ok := prediction.TotalTime(); ok {
			apiPrediction.TotalTime = nullable.NewNullableWithValue(totalTime.String())
		}
		if laps, ok := prediction.Laps(); ok {
			apiPrediction.Laps = nullable.NewNullableWithValue(laps)
		}
		body.Prediction = nullable.NewNullableWithValue(apiPrediction)
	}
	return body
}
