// Package presentation wires the HTTP routes of the backend to the controllers.
package presentation

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/way-station/tasuki/app/backend/internal/presentation/api"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

// Controllers holds every controller the router serves.
type Controllers struct {
	Health        *controller.HealthController
	Auth          *controller.AuthController
	LineLogin     *controller.LineLoginController
	Races         *controller.RaceController
	Participation *controller.ParticipationController
	RunningOrders *controller.RunningOrderController
	Laps          *controller.LapController
	Roster        *controller.RosterController
}

// Options configures the router.
type Options struct {
	// Authenticate identifies the caller of the API routes that need a session (ADR-0014).
	Authenticate func(http.Handler) http.Handler
	// ResolveCaller resolves the caller of the routes that check a role, after
	// Authenticate (docs/design/api.md).
	ResolveCaller func(http.Handler) http.Handler
	// FrontendOrigin is the only Origin allowed on API requests that change data (ADR-0014).
	FrontendOrigin string
	// DevLogin registers POST /api/v1/dev/login (ADR-0009). Never set in production.
	DevLogin bool
	// LineLogin registers POST /api/v1/auth/line/authorize and GET /api/v1/auth/line/callback
	// (ADR-0007). It is false while LINE Login is not configured.
	LineLogin bool
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

	apiHandler := newAPIHandler(Server{AuthController: controllers.Auth, LineLoginController: controllers.LineLogin, RaceController: controllers.Races, ParticipationController: controllers.Participation, RunningOrderController: controllers.RunningOrders, LapController: controllers.Laps, RosterController: controllers.Roster}, options.Logger)
	// The wrapper reads path parameters such as {race_id} before calling apiHandler.
	paramHandler := &api.ServerInterfaceWrapper{Handler: apiHandler, ErrorHandlerFunc: writeInvalidRequest}
	router.Route("/api/v1", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.RequireOrigin(options.FrontendOrigin))
		if options.DevLogin {
			apiRouter.Post("/dev/login", apiHandler.DevLogin)
		}
		apiRouter.Get("/invite-links/{token}", paramHandler.GetInviteLink)
		apiRouter.Post("/session", apiHandler.CreateSession)
		if options.LineLogin {
			apiRouter.Post("/auth/line/authorize", apiHandler.AuthorizeLine)
			apiRouter.Get("/auth/line/callback", paramHandler.LineCallback)
		}
		apiRouter.Group(func(userRouter chi.Router) {
			userRouter.Use(options.Authenticate)
			userRouter.Get("/me", apiHandler.GetMe)
			userRouter.Delete("/session", apiHandler.Logout)

			userRouter.Group(func(callerRouter chi.Router) {
				callerRouter.Use(options.ResolveCaller)
				organizer := callerRouter.With(middleware.RequireOrganizer)
				organizer.Get("/races/{race_id}/invitations", paramHandler.ListInvitations)
				organizer.Put("/races/{race_id}/invitations/{roster_member_id}", paramHandler.Invite)
				organizer.Delete("/races/{race_id}/invitations/{roster_member_id}", paramHandler.CancelInvitation)
				callerRouter.With(middleware.RequireMember).Put("/races/{race_id}/attendance", paramHandler.AnswerAttendance)
				organizer.Get("/races/{race_id}/entries", paramHandler.ListEntries)
				organizer.Post("/races/{race_id}/entries", paramHandler.CreateEntry)
				organizer.Put("/races/{race_id}/entries/{entry_id}/members/{roster_member_id}", paramHandler.AddEntryMember)
				organizer.Delete("/races/{race_id}/entries/{entry_id}/members/{roster_member_id}", paramHandler.RemoveEntryMember)
				callerRouter.With(middleware.RequireMember).Get("/races/{race_id}/running-orders", paramHandler.ListRunningOrders)
				organizer.Put("/races/{race_id}/entries/{entry_id}/running-order", paramHandler.SaveRunningOrder)
				organizer.Post("/races/{race_id}/entries/{entry_id}/running-order/confirm", paramHandler.ConfirmRunningOrder)
				callerRouter.With(middleware.RequireMember).Get("/races", apiHandler.ListRaces)
				callerRouter.With(middleware.RequireOrganizer).Post("/races", apiHandler.CreateRace)
				callerRouter.With(middleware.RequireMember).Get("/races/{race_id}", paramHandler.GetRace)
				callerRouter.With(middleware.RequireOrganizer).Put("/races/{race_id}", paramHandler.UpdateRace)
				member := callerRouter.With(middleware.RequireMember)
				member.Get("/races/{race_id}/my-laps", paramHandler.GetMyLaps)
				member.Post("/races/{race_id}/my-laps", paramHandler.SubmitLaps)
				member.Patch("/races/{race_id}/my-laps/{lap_id}", paramHandler.EditLap)
				member.Delete("/races/{race_id}/my-laps/{lap_id}", paramHandler.DeleteLap)
				member.Get("/me/profile", paramHandler.GetMyProfile)
				callerRouter.With(middleware.RequireOrganizer).Get("/races/{race_id}/laps", paramHandler.ListRaceLaps)
				organizer.Get("/roster-members", paramHandler.ListRosterMembers)
				organizer.Post("/roster-members", paramHandler.CreateRosterMember)
				organizer.Get("/roster-members/{roster_member_id}", paramHandler.GetRosterMember)
				organizer.Patch("/roster-members/{roster_member_id}", paramHandler.UpdateRosterMember)
				organizer.Delete("/roster-members/{roster_member_id}/link", paramHandler.UnlinkRosterMember)
				organizer.Post("/roster-members/{roster_member_id}/invite-links", paramHandler.CreateInviteLink)
				organizer.Get("/member-links", paramHandler.ListPendingMemberLinks)
				organizer.Post("/member-links/{link_id}/approve", paramHandler.ApproveMemberLink)
				organizer.Post("/member-links/{link_id}/reject", paramHandler.RejectMemberLink)
			})
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
