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
	"github.com/way-station/tasuki/app/backend/internal/application/laps"
	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/config"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/mongodb"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/query"
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
	if err := mongodb.EnsureIndexes(connectCtx, mongoClient); err != nil {
		_ = mongoClient.Close(connectCtx)
		return err
	}

	// The outbox runs the steps of the writes across PostgreSQL and MongoDB (ADR-0016).
	lapSteps := newLapStepRunner(postgresClient, mongoClient)
	outboxStore, err := postgres.NewOutbox(postgresClient, random.IDGenerator{}, time.Now, logger, postgres.LapOutboxSteps(lapSteps)...)
	if err != nil {
		_ = mongoClient.Close(connectCtx)
		return err
	}
	go outboxStore.RunWorker(ctx)

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
		Health: controller.NewHealthController(check),
		Auth:   authController,
		Laps:   newLaps(postgresClient, mongoClient, outboxStore, shared.TeamID(cfg.TeamID)),
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

// newLapStepRunner wires what runs the outbox steps of the laps: the lap records in MongoDB
// and the average times in PostgreSQL.
func newLapStepRunner(postgresClient *postgres.Client, mongoClient *mongodb.Client) *laps.LapStepRunner {
	ids := random.IDGenerator{}
	lapFactory := lap.NewFactory(ids)
	records := mongodb.NewLapRecordRepository(mongoClient, lapFactory, lap.RecordFactory{}, time.Now)
	members := postgres.NewRosterMemberRepository(postgresClient, roster.NewMemberFactory(ids))
	return laps.NewLapStepRunner(records, members, lapFactory, lap.RecordFactory{}, roster.AverageService{})
}

// newLaps wires the use cases of the lap times and the profile for the team. The writes add
// their steps to outboxStore.
func newLaps(postgresClient *postgres.Client, mongoClient *mongodb.Client, outboxStore outbox.Outbox, teamID shared.TeamID) *controller.LapController {
	ids := random.IDGenerator{}
	unitOfWork := postgres.NewUnitOfWork(postgresClient)
	records := mongodb.NewLapRecordRepository(mongoClient, lap.NewFactory(ids), lap.RecordFactory{}, time.Now)
	profiles := postgres.NewProfileQueryService(postgresClient)
	queries := query.NewLapQueryService(postgres.NewLapReader(postgresClient), mongodb.NewLapRecordReader(mongoClient), profiles)
	dtoFactory := laps.LapDTOFactory{}
	return controller.NewLapController(
		laps.NewGetMyLapsUseCase(queries, dtoFactory, teamID),
		laps.NewSubmitLapsUseCase(unitOfWork, queries, outboxStore, lap.SubmissionPolicy{}, lap.NewFactory(ids), outbox.EntryFactory{}, time.Now, teamID),
		laps.NewEditLapUseCase(unitOfWork, queries, records, outboxStore, lap.SubmissionPolicy{}, outbox.EntryFactory{}, teamID),
		laps.NewDeleteLapUseCase(unitOfWork, queries, records, outboxStore, lap.SubmissionPolicy{}, outbox.EntryFactory{}, teamID),
		laps.NewListRaceLapsUseCase(queries, lap.VisibilityPolicy{}, dtoFactory, teamID),
		laps.NewGetProfileUseCase(profiles, dtoFactory),
		laps.LapCommandFactory{},
	)
}
