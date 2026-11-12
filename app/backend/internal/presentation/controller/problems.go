package controller

import (
	"net/http"

	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

func unauthenticated() api.UnauthenticatedApplicationProblemPlusJSONResponse {
	return api.UnauthenticatedApplicationProblemPlusJSONResponse(middleware.UnauthenticatedProblem())
}

func notFound() api.NotFoundApplicationProblemPlusJSONResponse {
	return api.NotFoundApplicationProblemPlusJSONResponse{Type: "about:blank", Title: "Not Found", Status: http.StatusNotFound, Code: api.ProblemCodeNotFound}
}
