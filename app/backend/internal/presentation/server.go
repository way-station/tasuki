package presentation

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
)

// Server implements api.StrictServerInterface with the controllers. Operations without a
// controller fall through to notImplemented, which is embedded one level deeper.
type Server struct {
	*controller.AuthController
	fallback
}

type fallback struct {
	notImplemented
}

var _ api.StrictServerInterface = Server{}

// newAPIHandler wraps the server in the generated strict handler.
func newAPIHandler(server Server, logger *slog.Logger) api.ServerInterface {
	return api.NewStrictHandlerWithOptions(server, nil, api.StrictHTTPServerOptions{
		RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
			writeProblem(w, api.Problem{Type: "about:blank", Title: "Bad Request", Status: http.StatusBadRequest, Code: api.ProblemCodeInvalidRequest, Detail: ptr(err.Error())})
		},
		ResponseErrorHandlerFunc: func(w http.ResponseWriter, req *http.Request, err error) {
			if errors.Is(err, errNotImplemented) {
				http.Error(w, http.StatusText(http.StatusNotImplemented), http.StatusNotImplemented)
				return
			}
			logger.ErrorContext(req.Context(), "request failed", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		},
	})
}

func ptr[T any](value T) *T { return &value }

func writeProblem(w http.ResponseWriter, problem api.Problem) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(problem.Status)
	_ = json.NewEncoder(w).Encode(problem)
}
