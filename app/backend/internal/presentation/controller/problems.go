package controller

import (
	"errors"
	"net/http"

	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
)

// stateConflict returns the 409 problem for a shared.StateError.
func stateConflict(err error) (api.ConflictApplicationProblemPlusJSONResponse, bool) {
	var stateErr *shared.StateError
	if !errors.As(err, &stateErr) {
		return api.ConflictApplicationProblemPlusJSONResponse{}, false
	}
	detail := stateErr.Message
	return api.ConflictApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Conflict", Status: http.StatusConflict, Code: api.ProblemCodeStateConflict, Detail: &detail}, true
}

func notFound() api.NotFoundApplicationProblemPlusJSONResponse {
	return api.NotFoundApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound, Code: api.ProblemCodeNotFound}
}
