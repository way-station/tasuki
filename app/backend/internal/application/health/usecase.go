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
func (useCase *CheckUseCase) Execute(ctx context.Context) StatusDTO {
	var mongoErr, postgresErr error
	var waitGroup sync.WaitGroup
	waitGroup.Go(func() { mongoErr = useCase.mongo.PingMongo(ctx) })
	waitGroup.Go(func() { postgresErr = useCase.postgres.PingPostgres(ctx) })
	waitGroup.Wait()

	if mongoErr != nil {
		useCase.logger.ErrorContext(ctx, "health check failed", "store", "mongo", "error", mongoErr)
	}
	if postgresErr != nil {
		useCase.logger.ErrorContext(ctx, "health check failed", "store", "postgres", "error", postgresErr)
	}
	return useCase.factory.Create(mongoErr == nil, postgresErr == nil)
}
