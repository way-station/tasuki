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
	"github.com/way-station/tasuki/app/backend/internal/config"
	"github.com/way-station/tasuki/app/backend/internal/domain/identity"
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

	check := health.NewCheckUseCase(mongodb.NewHealthQueryService(mongoClient), postgres.NewHealthQueryService(postgresClient), health.StatusDTOFactory{}, logger)
	authenticate, authController := newAuth(cfg, postgresClient)
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
	}, presentation.Options{
		Authenticate: middleware.Authenticate(authenticate, logger),
		DevLogin:     cfg.DevLoginEnabled,
		Frontend:     frontend,
		Logger:       logger,
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
func newAuth(cfg config.Config, postgresClient *postgres.Client) (*auth.AuthenticateUseCase, *controller.AuthController) {
	ids := random.IDGenerator{}
	accountFactory := identity.NewAccountFactory(ids)
	sessionFactory := identity.NewSessionFactory(ids, random.TokenGenerator{})
	sessions := postgres.NewSessionRepository(postgresClient, sessionFactory)

	authenticate := auth.NewAuthenticateUseCase(postgres.NewSessionQueryService(postgresClient), time.Now)
	login := auth.NewLoginUseCase(postgres.NewAccountRepository(postgresClient, accountFactory), sessions, accountFactory, sessionFactory, auth.SessionDTOFactory{}, time.Now)
	logout := auth.NewLogoutUseCase(sessions, time.Now)
	getMe := auth.NewGetMeUseCase(postgres.NewMeQueryService(postgresClient), accountFactory, identity.NewMemberLinkFactory(ids),
		identity.LinkStatusService{}, identity.NewOrganizerPolicy(cfg.OrganizerLineUserIDs), auth.MeDTOFactory{})
	return authenticate, controller.NewAuthController(login, logout, getMe, auth.LoginCommandFactory{})
}
