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
	"github.com/way-station/tasuki/app/backend/internal/application/participations"
	"github.com/way-station/tasuki/app/backend/internal/application/rosters"
	"github.com/way-station/tasuki/app/backend/internal/application/runningorders"
	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/config"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
	"github.com/way-station/tasuki/app/backend/internal/domain/lap"
	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/race"
	"github.com/way-station/tasuki/app/backend/internal/domain/roster"
	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
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

	// "Today" is a date in Japan, for example when a leaving member's invitations to races
	// dated today or later are cancelled (docs/design/api.md).
	tokyo, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
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
		Health: controller.NewHealthController(check),
		Auth:   authController,
		Roster: newRoster(postgresClient, mongoClient, outboxStore, shared.TeamID(cfg.TeamID), cfg.FrontendOrigin, tokyo),
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

// newRoster wires the use cases of the roster and the approval of member links for the team.
// Marking a member as left adds the steps that delete their laps to outboxStore.
func newRoster(postgresClient *postgres.Client, mongoClient *mongodb.Client, outboxStore outbox.Outbox, teamID shared.TeamID, frontendOrigin string, location *time.Location) *controller.RosterController {
	ids := random.IDGenerator{}
	unitOfWork := postgres.NewUnitOfWork(postgresClient)
	memberFactory := roster.NewMemberFactory(ids)
	members := postgres.NewRosterMemberRepository(postgresClient, memberFactory)
	linkFactory := identity.NewMemberLinkFactory(ids)
	links := postgres.NewMemberLinkRepository(postgresClient, linkFactory)
	inviteLinkFactory := identity.NewInviteLinkFactory(ids, random.TokenGenerator{})
	sessions := postgres.NewSessionRepository(postgresClient, identity.NewSessionFactory(ids, random.TokenGenerator{}))
	queries := query.NewRosterQueryService(postgres.NewRosterReader(postgresClient), mongodb.NewMemberLapReader(mongoClient))
	invitations := newInvitationCanceller(postgresClient, unitOfWork, outboxStore, teamID)
	dtoFactory := rosters.RosterDTOFactory{}
	return controller.NewRosterController(
		rosters.NewListRosterMembersUseCase(queries, linkFactory, inviteLinkFactory, identity.LinkStatusService{}, dtoFactory, time.Now, teamID),
		rosters.NewAddRosterMemberUseCase(members, memberFactory, dtoFactory, teamID),
		rosters.NewGetRosterMemberUseCase(queries, dtoFactory, teamID),
		rosters.NewSetLeftUseCase(unitOfWork, members, queries, invitations, links, sessions, sessions, roster.LeaveService{}, time.Now, location, teamID),
		rosters.NewUnlinkUseCase(unitOfWork, members, links, time.Now, teamID),
		rosters.NewCreateInviteLinkUseCase(unitOfWork, members, links, postgres.NewInviteLinkRepository(postgresClient, inviteLinkFactory), inviteLinkFactory, identity.LinkStatusService{}, dtoFactory, time.Now, teamID),
		rosters.NewListPendingLinksUseCase(queries, dtoFactory, teamID),
		rosters.NewDecideLinkUseCase(unitOfWork, members, links, time.Now, teamID),
		rosters.RosterCommandFactory{},
		frontendOrigin,
	)
}

// newInvitationCanceller wires what cancels the invitations of a member who leaves: the
// member leaves the entries and the running orders, and loses their laps of a confirmed
// running order through outboxStore. It joins the unit of work of marking them as left.
func newInvitationCanceller(postgresClient *postgres.Client, unitOfWork transaction.UnitOfWork, outboxStore outbox.Outbox, teamID shared.TeamID) *participations.LeaveInvitationCanceller {
	ids := random.IDGenerator{}
	entries := postgres.NewEntryRepository(postgresClient, participation.NewEntryFactory(ids))
	runningOrders := postgres.NewRunningOrderRepository(postgresClient, runningorder.NewFactory(ids))
	runners := runningorders.NewRemoveRunnerUseCase(entries, runningOrders, laps.NewRemoveLapsUseCase(outboxStore, outbox.EntryFactory{}))
	cancel := participations.NewCancelInvitationUseCase(unitOfWork, postgres.NewRaceRepository(postgresClient, race.NewFactory(ids)),
		postgres.NewInvitationRepository(postgresClient, participation.InvitationFactory{}), entries, runners, teamID)
	return participations.NewLeaveInvitationCanceller(cancel, participations.ParticipationCommandFactory{})
}
