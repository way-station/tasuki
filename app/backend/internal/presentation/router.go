// Package presentation wires the HTTP routes of the backend to the controllers.
package presentation

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
)

// Controllers holds every controller the router serves.
type Controllers struct {
	Health *controller.HealthController
}

// NewRouter returns the HTTP handler with all backend routes registered.
func NewRouter(c Controllers) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Get("/health", c.Health.Get)
	return r
}
