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
// When frontend is not nil, it serves every GET and HEAD outside /api and /health, so
// the API and the frontend share one origin (ADR-0018). In development it is
// nil, and the Vite dev server serves the frontend.
func NewRouter(controllers Controllers, frontend http.Handler) http.Handler {
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.Recoverer)

	router.Get("/health", controllers.Health.Get)

	if frontend != nil {
		// An unknown API path is a 404, never the frontend.
		router.Handle("/api/*", http.NotFoundHandler())
		router.Get("/*", frontend.ServeHTTP)
		router.Head("/*", frontend.ServeHTTP)
	}
	return router
}
