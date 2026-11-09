// Command server runs the tasuki backend API.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/way-station/tasuki/app/backend/internal/application/auth"
	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/application/participations"
	"github.com/way-station/tasuki/app/backend/internal/config"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/mongodb"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
	"github.com/way-station/tasuki/app/backend/internal/presentation"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
	"github.com/way-station/tasuki/app/backend/internal/presentation/middleware"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := run(logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	mongoClient, err := mongodb.Connect(connectCtx, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		return err
	}
	postgresClient, err := postgres.Connect(connectCtx, cfg.DatabaseURL)
	if err != nil {
		_ = mongoClient.Close(connectCtx)
		return err
	}
	defer postgresClient.Close()
	if err := postgres.RequireTeam(connectCtx, postgresClient, shared.TeamID(cfg.TeamID)); err != nil {
		_ = mongoClient.Close(connectCtx)
		return err
	}

	check := health.NewCheckUseCase(mongodb.NewHealthQueryService(mongoClient), postgres.NewHealthQueryService(postgresClient), health.StatusDTOFactory{}, logger)
	authenticate, getMe, authController := newAuth(cfg, postgresClient)
	var frontend http.Handler
	if cfg.FrontendDir != "" {
		frontend = presentation.NewFrontendHandler(cfg.FrontendDir)
	}
	if cfg.DevLoginEnabled {
		logger.Warn("DEV-ONLY LOGIN IS ENABLED: anyone can log in as any LINE user. Never enable it in production (ADR-0009).")
	}
	router := presentation.NewRouter(presentation.Controllers{
		Health:        controller.NewHealthController(check),
		Auth:          authController,
		Participation: newParticipation(postgresClient, shared.TeamID(cfg.TeamID)),
	}, presentation.Options{
		Authenticate:   middleware.Authenticate(authenticate, logger),
		ResolveCaller:  middleware.ResolveCaller(getMe, logger),
		FrontendOrigin: cfg.FrontendOrigin,
		DevLogin:       cfg.DevLoginEnabled,
		Frontend:       frontend,
		Logger:         logger,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("server listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
		close(errCh)
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return mongoClient.Close(shutdownCtx)
}

// newAuth wires the use cases of logging in and identifying the caller.
func newAuth(cfg config.Config, postgresClient *postgres.Client) (*auth.AuthenticateUseCase, *auth.GetMeUseCase, *controller.AuthController) {
	ids := random.IDGenerator{}
	userFactory := identity.NewUserFactory(ids)
	sessionFactory := identity.NewSessionFactory(ids, random.TokenGenerator{})
	sessions := postgres.NewSessionRepository(postgresClient, sessionFactory)

	authenticate := auth.NewAuthenticateUseCase(postgres.NewSessionQueryService(postgresClient), time.Now)
	login := auth.NewLoginUseCase(postgres.NewUserRepository(postgresClient, userFactory), sessions, userFactory, sessionFactory, auth.SessionDTOFactory{}, time.Now)
	logout := auth.NewLogoutUseCase(sessions, time.Now)
	getMe := auth.NewGetMeUseCase(postgres.NewMeQueryService(postgresClient), userFactory, identity.NewMemberLinkFactory(ids),
		identity.LinkStatusService{}, identity.NewOrganizerPolicy(cfg.OrganizerLineUserIDs), auth.MeDTOFactory{})
	return authenticate, getMe, controller.NewAuthController(login, logout, getMe, auth.LoginCommandFactory{})
}

// newParticipation wires the use cases of invitations, attendance, and entries for the team.
func newParticipation(postgresClient *postgres.Client, teamID shared.TeamID) *controller.ParticipationController {
	ids := random.IDGenerator{}
	unitOfWork := postgres.NewUnitOfWork(postgresClient)
	races := postgres.NewRaceRepository(postgresClient, race.NewFactory(ids))
	members := postgres.NewRosterMemberRepository(postgresClient, roster.NewMemberFactory(ids))
	invitations := postgres.NewInvitationRepository(postgresClient, participation.InvitationFactory{})
	attendances := postgres.NewAttendanceRepository(postgresClient, participation.AttendanceFactory{})
	entryFactory := participation.NewEntryFactory(ids)
	entries := postgres.NewEntryRepository(postgresClient, entryFactory)
	dtoFactory := participations.ParticipationDTOFactory{}
	return controller.NewParticipationController(controller.ParticipationUseCases{
		ListInvitations:   participations.NewListInvitationsUseCase(postgres.NewInvitationQueryService(postgresClient), participation.AttendancePolicy{}, dtoFactory, teamID),
		Invite:            participations.NewInviteUseCase(races, members, invitations, participation.InvitationFactory{}, teamID),
		CancelInvitation:  participations.NewCancelInvitationUseCase(unitOfWork, races, invitations, entries, noRunners{}, teamID),
		AnswerAttendance:  participations.NewAnswerAttendanceUseCase(unitOfWork, invitations, attendances, entries, noRunners{}, participation.AttendanceFactory{}, dtoFactory),
		ListEntries:       participations.NewListEntriesUseCase(postgres.NewEntryQueryService(postgresClient), dtoFactory, teamID),
		CreateEntry:       participations.NewCreateEntryUseCase(races, entries, entryFactory, dtoFactory, teamID),
		AddEntryMember:    participations.NewAddEntryMemberUseCase(races, entries, attendances, teamID),
		RemoveEntryMember: participations.NewRemoveEntryMemberUseCase(unitOfWork, races, entries, noRunners{}, teamID),
	}, participations.ParticipationCommandFactory{})
}

// noRunners stands in for the running order (#95) and lap (#99) Application layers, which
// take a member off the running orders when they leave an entry. No running order can be
// saved before them, so there is no runner to take off.
type noRunners struct{}

func (noRunners) RemoveRunner(context.Context, shared.RaceID, shared.MemberID) error {
	return nil
}
