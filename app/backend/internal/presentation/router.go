// Package presentation wires the HTTP routes of the backend to the controllers.
package presentation

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

// Controllers holds every controller the router serves.
type Controllers struct {
	Health *controller.HealthController
	Auth   *controller.AuthController
}

// Options configures the router.
type Options struct {
	// Authenticate identifies the caller of the API routes that need a session (ADR-0014).
	Authenticate func(http.Handler) http.Handler
	// FrontendOrigin is the only Origin allowed on API requests that change data (ADR-0014).
	FrontendOrigin string
	// DevLogin registers POST /api/v1/dev/login (ADR-0009). Never set in production.
	DevLogin bool
	// Frontend serves every GET and HEAD outside /api and /health, so the API and the
	// frontend share one origin (ADR-0018). In development it is nil, and the Vite dev
	// server serves the frontend.
	Frontend http.Handler
	Logger   *slog.Logger
}

// NewRouter returns the HTTP handler with all backend routes registered.
func NewRouter(controllers Controllers, options Options) http.Handler {
	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.Recoverer)

	router.Get("/health", controllers.Health.Get)

	apiHandler := newAPIHandler(Server{AuthController: controllers.Auth}, options.Logger)
	router.Route("/api/v1", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.RequireOrigin(options.FrontendOrigin))
		if options.DevLogin {
			apiRouter.Post("/dev/login", apiHandler.DevLogin)
		}
		apiRouter.Group(func(userRouter chi.Router) {
			userRouter.Use(options.Authenticate)
			userRouter.Get("/me", apiHandler.GetMe)
			userRouter.Delete("/session", apiHandler.Logout)
		})
	})

	if options.Frontend != nil {
		// An unknown API path is a 404, never the frontend.
		router.Handle("/api/*", http.NotFoundHandler())
		router.Get("/*", options.Frontend.ServeHTTP)
		router.Head("/*", options.Frontend.ServeHTTP)
	}
	return router
}
