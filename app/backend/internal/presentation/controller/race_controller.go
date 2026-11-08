package controller

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/oapi-codegen/nullable"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/way-station/tasuki/app/backend/internal/application/races"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

// RaceController serves the race operations (docs/design/api.md). The routes check the
// caller's role before it runs; reads apply what the caller may see in the use cases.
type RaceController struct {
	list           *races.ListRacesUseCase
	get            *races.GetRaceUseCase
	create         *races.CreateRaceUseCase
	update         *races.UpdateRaceUseCase
	commandFactory races.RaceCommandFactory
}

// NewRaceController returns a RaceController.
func NewRaceController(list *races.ListRacesUseCase, get *races.GetRaceUseCase, create *races.CreateRaceUseCase, update *races.UpdateRaceUseCase, commandFactory races.RaceCommandFactory) *RaceController {
	return &RaceController{list: list, get: get, create: create, update: update, commandFactory: commandFactory}
}

// ListRaces returns the races the caller may see, newest registered first.
func (controller *RaceController) ListRaces(ctx context.Context, _ api.ListRacesRequestObject) (api.ListRacesResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.ListRaces401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	summaries, err := controller.list.Execute(ctx, caller)
	if err != nil {
		return nil, err
	}
	response := api.ListRaces200JSONResponse{Races: make([]api.RaceSummary, 0, len(summaries))}
	for _, summary := range summaries {
		raceDate, err := toAPIDate(summary.RaceDate())
		if err != nil {
			return nil, err
		}
		item := api.RaceSummary{Id: string(summary.ID()), Name: summary.Name(), RaceDate: raceDate, MyAttendance: toAPIAttendance(summary.MyAttendance())}
		if confirmed, visible := summary.RunningOrderConfirmed(); visible {
			item.RunningOrderConfirmed = &confirmed
		}
		response.Races = append(response.Races, item)
	}
	return response, nil
}

// CreateRace registers a race and returns it.
func (controller *RaceController) CreateRace(ctx context.Context, request api.CreateRaceRequestObject) (api.CreateRaceResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.CreateRace401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	if request.Body == nil {
		return api.CreateRace400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	command, err := controller.commandFactory.CreateRace(request.Body.Name, toDescriptionInput(request.Body.Description))
	if problem, ok := validationFailed(err); ok {
		return api.CreateRace422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: problem}, nil
	}
	if err != nil {
		return nil, err
	}
	raceID, err := controller.create.Execute(ctx, command)
	if problem, ok := validationFailed(err); ok {
		return api.CreateRace422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: problem}, nil
	}
	if err != nil {
		return nil, err
	}
	created, err := controller.get.Execute(ctx, caller, raceID)
	if err != nil {
		return nil, err
	}
	body, err := toAPIRace(created)
	if err != nil {
		return nil, err
	}
	return api.CreateRace201JSONResponse(body), nil
}

// GetRace returns a race as the caller may see it. A race the caller may not see is 404.
func (controller *RaceController) GetRace(ctx context.Context, request api.GetRaceRequestObject) (api.GetRaceResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.GetRace401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	found, err := controller.get.Execute(ctx, caller, shared.RaceID(request.RaceId))
	if errors.Is(err, shared.ErrNotFound) {
		return api.GetRace404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	}
	if err != nil {
		return nil, err
	}
	body, err := toAPIRace(found)
	if err != nil {
		return nil, err
	}
	return api.GetRace200JSONResponse(body), nil
}

// UpdateRace saves an edit and returns the race.
func (controller *RaceController) UpdateRace(ctx context.Context, request api.UpdateRaceRequestObject) (api.UpdateRaceResponseObject, error) {
	caller, ok := middleware.Caller(ctx)
	if !ok {
		return api.UpdateRace401ApplicationProblemPlusJSONResponse{UnauthenticatedApplicationProblemPlusJSONResponse: unauthenticated()}, nil
	}
	if request.Body == nil {
		return api.UpdateRace400ApplicationProblemPlusJSONResponse{BadRequestApplicationProblemPlusJSONResponse: badRequest("a body is required")}, nil
	}
	raceID := shared.RaceID(request.RaceId)
	command, err := controller.commandFactory.UpdateRace(raceID, request.Body.Name, toDescriptionInput(request.Body.Description), request.Body.Version)
	if problem, ok := validationFailed(err); ok {
		return api.UpdateRace422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: problem}, nil
	}
	if err != nil {
		return nil, err
	}
	err = controller.update.Execute(ctx, command)
	if problem, ok := validationFailed(err); ok {
		return api.UpdateRace422ApplicationProblemPlusJSONResponse{ValidationFailedApplicationProblemPlusJSONResponse: problem}, nil
	}
	switch {
	case errors.Is(err, shared.ErrNotFound):
		return api.UpdateRace404ApplicationProblemPlusJSONResponse{NotFoundApplicationProblemPlusJSONResponse: notFound()}, nil
	case errors.Is(err, shared.ErrVersionConflict):
		return api.UpdateRace409ApplicationProblemPlusJSONResponse{ConflictApplicationProblemPlusJSONResponse: versionConflict()}, nil
	case err != nil:
		return nil, err
	}
	updated, err := controller.get.Execute(ctx, caller, raceID)
	if err != nil {
		return nil, err
	}
	body, err := toAPIRace(updated)
	if err != nil {
		return nil, err
	}
	return api.UpdateRace200JSONResponse(body), nil
}

func toDescriptionInput(description api.RaceDescription) races.DescriptionInput {
	input := races.DescriptionInput{
		Type:     string(description.Type),
		RaceDate: description.RaceDate.Format(time.DateOnly),
	}
	if description.Deadline != nil {
		input.Deadline = *description.Deadline
	}
	if description.Text != nil {
		input.Text = *description.Text
	}
	if description.LegDistancesKm != nil {
		input.LegDistancesKm = make([]float64, 0, len(*description.LegDistancesKm))
		for _, distance := range *description.LegDistancesKm {
			input.LegDistancesKm = append(input.LegDistancesKm, toFloat64(distance))
		}
	}
	if description.TimeLimitMinutes != nil {
		input.TimeLimitMinutes = *description.TimeLimitMinutes
	}
	if description.LapDistanceKm != nil {
		input.LapDistanceKm = toFloat64(*description.LapDistanceKm)
	}
	if description.Target != nil {
		input.TargetTotalTime = description.Target.TotalTime
		input.TargetLaps = description.Target.Laps
	}
	return input
}

func toAPIRace(dto races.RaceDTO) (api.Race, error) {
	description := dto.Description()
	raceDate, err := toAPIDate(description.RaceDate())
	if err != nil {
		return api.Race{}, err
	}
	apiDescription := api.RaceDescription{
		Type:     api.RaceType(description.Type()),
		RaceDate: raceDate,
		Text:     new(description.Text()),
	}
	if deadline, ok := description.Deadline(); ok {
		apiDescription.Deadline = &deadline
	}
	switch description.Type() {
	case race.Distance:
		legs := make([]float32, 0, len(description.LegDistancesKm()))
		for _, distance := range description.LegDistancesKm() {
			legs = append(legs, float32(distance))
		}
		apiDescription.LegDistancesKm = &legs
		if target, ok := description.TargetTotalTime(); ok {
			apiDescription.Target = &api.RaceTarget{TotalTime: new(target.String())}
		}
	case race.Time:
		apiDescription.TimeLimitMinutes = new(description.TimeLimitMinutes())
		apiDescription.LapDistanceKm = new(float32(description.LapDistanceKm()))
		if laps, ok := description.TargetLaps(); ok {
			apiDescription.Target = &api.RaceTarget{Laps: &laps}
		}
	}

	body := api.Race{
		Id:            string(dto.ID()),
		Name:          dto.Name(),
		Description:   apiDescription,
		MyAttendance:  toAPIAttendance(dto.MyAttendance()),
		CanSubmitLaps: dto.CanSubmitLaps(),
	}
	if version, ok := dto.Version(); ok {
		body.Version = &version
	}
	if confirmed, visible := dto.RunningOrderConfirmed(); visible {
		body.RunningOrderConfirmed = &confirmed
	}
	return body, nil
}

// toAPIAttendance returns my_attendance: omitted when the caller is not invited, and null
// when they have not answered (docs/design/api.md).
func toAPIAttendance(answer race.Answer, invited bool) nullable.Nullable[api.MyAttendance] {
	if !invited {
		return nullable.Nullable[api.MyAttendance]{}
	}
	switch answer {
	case race.Attending:
		return nullable.NewNullableWithValue(api.MyAttendanceAttending)
	case race.Undecided:
		return nullable.NewNullableWithValue(api.MyAttendanceUndecided)
	case race.NotAttending:
		return nullable.NewNullableWithValue(api.MyAttendanceNotAttending)
	}
	return nullable.NewNullNullable[api.MyAttendance]()
}

func toAPIDate(date shared.Date) (openapi_types.Date, error) {
	parsed, err := time.Parse(time.DateOnly, date.String())
	if err != nil {
		return openapi_types.Date{}, err
	}
	return openapi_types.Date{Time: parsed}, nil
}

// toFloat64 widens a float32 of the request through its shortest decimal form, so that
// 3.2 stays 3.2 instead of becoming 3.2000000476837.
func toFloat64(value float32) float64 {
	widened, _ := strconv.ParseFloat(strconv.FormatFloat(float64(value), 'g', -1, 32), 64)
	return widened
}

// validationFailed returns the 422 problem for a shared.ValidationError.
func validationFailed(err error) (api.ValidationFailedApplicationProblemPlusJSONResponse, bool) {
	var validationErr *shared.ValidationError
	if !errors.As(err, &validationErr) {
		return api.ValidationFailedApplicationProblemPlusJSONResponse{}, false
	}
	fields := []struct {
		Field   string `json:"field"`
		Message string `json:"message"`
	}{{Field: validationErr.Field, Message: validationErr.Message}}
	return api.ValidationFailedApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Validation Failed", Status: http.StatusUnprocessableEntity, Code: api.ProblemCodeValidationFailed, Errors: &fields}, true
}

func unauthenticated() api.UnauthenticatedApplicationProblemPlusJSONResponse {
	return api.UnauthenticatedApplicationProblemPlusJSONResponse(middleware.UnauthenticatedProblem())
}

func notFound() api.NotFoundApplicationProblemPlusJSONResponse {
	return api.NotFoundApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound, Code: api.ProblemCodeNotFound}
}

func versionConflict() api.ConflictApplicationProblemPlusJSONResponse {
	return api.ConflictApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Conflict", Status: http.StatusConflict, Code: api.ProblemCodeVersionConflict}
}
