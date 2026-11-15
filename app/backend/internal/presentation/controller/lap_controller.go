package controller

import (
	"context"
	"errors"
	"net/http"

	"github.com/way-station/tasuki/app/backend/internal/application/laps"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

// LapController serves the lap times and the profile (docs/design/api.md). The routes check
// the caller's role before it runs; the use cases check that the caller is a runner in a
// confirmed running order.
type LapController struct {
	getMyLaps      *laps.GetMyLapsUseCase
	submit         *laps.SubmitLapsUseCase
	edit           *laps.EditLapUseCase
	remove         *laps.DeleteLapUseCase
	listRace       *laps.ListRaceLapsUseCase
	getProfile     *laps.GetProfileUseCase
	commandFactory laps.LapCommandFactory
}

// NewLapController returns a LapController.
func NewLapController(getMyLaps *laps.GetMyLapsUseCase, submit *laps.SubmitLapsUseCase, edit *laps.EditLapUseCase, remove *laps.DeleteLapUseCase, listRace *laps.ListRaceLapsUseCase, getProfile *laps.GetProfileUseCase, commandFactory laps.LapCommandFactory) *LapController {
	return &LapController{getMyLaps: getMyLaps, submit: submit, edit: edit, remove: remove, listRace: listRace, getProfile: getProfile, commandFactory: commandFactory}
}

// GetMyLaps returns the caller's laps for the race. A caller who is not a runner in a
// confirmed running order of the race is 404.
func (controller *LapController) GetMyLaps(ctx context.Context, request api.GetMyLapsRequestObject) (api.GetMyLapsResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.GetMyLaps401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	myLaps, err := controller.getMyLaps.Execute(ctx, caller, shared.RaceID(request.RaceId))
	if isLapNotFound(err) {
		return api.GetMyLaps404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	return api.GetMyLaps200JSONResponse(toAPIMyLaps(myLaps)), nil
}

// SubmitLaps submits the caller's laps for the race and returns all of their laps. A time
// that is not MM:SS is 400.
func (controller *LapController) SubmitLaps(ctx context.Context, request api.SubmitLapsRequestObject) (api.SubmitLapsResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.SubmitLaps401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	if request.Body == nil {
		return api.SubmitLaps400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	raceID := shared.RaceID(request.RaceId)
	command, err := controller.commandFactory.Submit(raceID, request.Body.Times)
	if err != nil {
		return api.SubmitLaps400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err.Error())}, nil
	}
	err = controller.submit.Execute(ctx, caller, command)
	if isLapNotFound(err) {
		return api.SubmitLaps404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	myLaps, err := controller.getMyLaps.Execute(ctx, caller, raceID)
	if err != nil {
		return nil, err
	}
	return api.SubmitLaps201JSONResponse(toAPIMyLaps(myLaps)), nil
}

// EditLap sets the time of one of the caller's laps and returns all of their laps. A lap
// the caller does not have in the race is 404.
func (controller *LapController) EditLap(ctx context.Context, request api.EditLapRequestObject) (api.EditLapResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.EditLap401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	if request.Body == nil {
		return api.EditLap400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	raceID := shared.RaceID(request.RaceId)
	command, err := controller.commandFactory.Edit(raceID, shared.LapID(request.LapId), request.Body.Time)
	if err != nil {
		return api.EditLap400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest(err.Error())}, nil
	}
	err = controller.edit.Execute(ctx, caller, command)
	if isLapNotFound(err) {
		return api.EditLap404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	myLaps, err := controller.getMyLaps.Execute(ctx, caller, raceID)
	if err != nil {
		return nil, err
	}
	return api.EditLap200JSONResponse(toAPIMyLaps(myLaps)), nil
}

// DeleteLap deletes one of the caller's laps and returns the rest. A lap the caller does
// not have in the race is 404.
func (controller *LapController) DeleteLap(ctx context.Context, request api.DeleteLapRequestObject) (api.DeleteLapResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.DeleteLap401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	raceID := shared.RaceID(request.RaceId)
	err := controller.remove.Execute(ctx, caller, controller.commandFactory.Delete(raceID, shared.LapID(request.LapId)))
	if isLapNotFound(err) {
		return api.DeleteLap404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	myLaps, err := controller.getMyLaps.Execute(ctx, caller, raceID)
	if err != nil {
		return nil, err
	}
	return api.DeleteLap200JSONResponse(toAPIMyLaps(myLaps)), nil
}

// ListRaceLaps returns the laps of every runner in the race. A race of another team is 404.
func (controller *LapController) ListRaceLaps(ctx context.Context, request api.ListRaceLapsRequestObject) (api.ListRaceLapsResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.ListRaceLaps401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	runners, err := controller.listRace.Execute(ctx, caller, shared.RaceID(request.RaceId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.ListRaceLaps404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	response := api.ListRaceLaps200JSONResponse{Runners: make([]api.RunnerLaps, 0, len(runners))}
	for _, runner := range runners {
		response.Runners = append(response.Runners, api.RunnerLaps{RosterMemberId: string(runner.MemberID()), Name: runner.Name(), Laps: toAPILaps(runner.Laps())})
	}
	return response, nil
}

// GetMyProfile returns the caller's name, average time, and lap count. An organizer without
// a roster member is not a member, so it is 403 ("Who can call" in docs/design/api.md).
func (controller *LapController) GetMyProfile(ctx context.Context, _ api.GetMyProfileRequestObject) (api.GetMyProfileResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.GetMyProfile401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	if _, linked := caller.RosterMember(); !linked {
		detail := "the caller's role does not allow this endpoint"
		return api.GetMyProfile403ApplicationProblemPlusJSONResponse{ForbiddenApplicationProblemPlusJSONResponse: api.ForbiddenApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Forbidden", Status: http.StatusForbidden, Code: api.ProblemCodeForbidden, Detail: &detail}}, nil
	}
	profile, err := controller.getProfile.Execute(ctx, caller)
	if err != nil {
		return nil, err
	}
	return api.GetMyProfile200JSONResponse{Name: profile.Name(), AverageTime: profile.Average().String(), LapCount: profile.LapCount()}, nil
}

// isLapNotFound reports whether err is answered with 404: the caller is not a runner in a
// confirmed running order of the race, or has no such lap.
func isLapNotFound(err error) bool {
	return errors.Is(err, lap.ErrNotRunner) || errors.Is(err, shared.ErrNotFound)
}

func toAPIMyLaps(dto laps.MyLapsDTO) api.MyLaps {
	return api.MyLaps{Laps: toAPILaps(dto.Laps()), AverageTime: dto.Average().String(), LapCount: dto.LapCount()}
}

func toAPILaps(dtos []laps.LapDTO) []api.Lap {
	apiLaps := make([]api.Lap, 0, len(dtos))
	for _, dto := range dtos {
		apiLaps = append(apiLaps, api.Lap{Id: string(dto.ID()), Time: dto.Time().String()})
	}
	return apiLaps
}
