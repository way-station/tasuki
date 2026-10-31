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

	"github.com/way-station/tasuki/app/backend/internal/application/health"
	"github.com/way-station/tasuki/app/backend/internal/config"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/mongodb"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres"
	"github.com/way-station/tasuki/app/backend/internal/presentation"
	"github.com/way-station/tasuki/app/backend/internal/presentation/controller"
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
	var frontend http.Handler
	if cfg.FrontendDir != "" {
		frontend = presentation.NewFrontendHandler(cfg.FrontendDir)
	}
	router := presentation.NewRouter(presentation.Controllers{
		Health: controller.NewHealthController(check),
	}, frontend)

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
