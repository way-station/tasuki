package health

import (
	"context"
	"log/slog"
	"sync"
)

// CheckUseCase checks whether the stores the backend depends on are reachable.
type CheckUseCase struct {
	mongo    MongoQueryService
	postgres PostgresQueryService
	factory  StatusDTOFactory
	logger   *slog.Logger
}

// NewCheckUseCase returns a CheckUseCase that reads through mongo and postgres.
func NewCheckUseCase(mongo MongoQueryService, postgres PostgresQueryService, factory StatusDTOFactory, logger *slog.Logger) *CheckUseCase {
	return &CheckUseCase{mongo: mongo, postgres: postgres, factory: factory, logger: logger}
}

// Execute runs the health check. A store failure is reported in the result, not as an error.
// The stores are pinged at the same time, so a slow store does not use up the other's deadline.
func (u *CheckUseCase) Execute(ctx context.Context) StatusDTO {
	var mongoErr, postgresErr error
	var wg sync.WaitGroup
	wg.Go(func() { mongoErr = u.mongo.PingMongo(ctx) })
	wg.Go(func() { postgresErr = u.postgres.PingPostgres(ctx) })
	wg.Wait()

	if mongoErr != nil {
		u.logger.ErrorContext(ctx, "health check failed", "store", "mongo", "error", mongoErr)
	}
	if postgresErr != nil {
		u.logger.ErrorContext(ctx, "health check failed", "store", "postgres", "error", postgresErr)
	}
	return u.factory.Create(mongoErr == nil, postgresErr == nil)
}
