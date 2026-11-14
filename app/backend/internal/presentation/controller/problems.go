package controller

import (
	"errors"
	"net/http"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

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

// stateConflict returns the 409 problem for a shared.StateError.
func stateConflict(err error) (api.ConflictApplicationProblemPlusJSONResponse, bool) {
	var stateErr *shared.StateError
	if !errors.As(err, &stateErr) {
		return api.ConflictApplicationProblemPlusJSONResponse{}, false
	}
	detail := stateErr.Message
	return api.ConflictApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Conflict", Status: http.StatusConflict, Code: api.ProblemCodeStateConflict, Detail: &detail}, true
}

func unauthenticated() api.UnauthenticatedApplicationProblemPlusJSONResponse {
	return api.UnauthenticatedApplicationProblemPlusJSONResponse(middleware.UnauthenticatedProblem())
}

func notFound() api.NotFoundApplicationProblemPlusJSONResponse {
	return api.NotFoundApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound, Code: api.ProblemCodeNotFound}
}
