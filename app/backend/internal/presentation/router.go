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
func NewRouter(controllers Controllers) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	router.Get("/health", controllers.Health.Get)
	return router
}
