package health

import (
	"context"
	"log/slog"
)

// CheckUseCase checks whether the stores the backend depends on are reachable.
type CheckUseCase struct {
	query   QueryService
	factory StatusDTOFactory
	logger  *slog.Logger
}

// NewCheckUseCase returns a CheckUseCase that reads through query.
func NewCheckUseCase(query QueryService, factory StatusDTOFactory, logger *slog.Logger) *CheckUseCase {
	return &CheckUseCase{query: query, factory: factory, logger: logger}
}

// Execute runs the health check. A store failure is reported in the result, not as an error.
func (u *CheckUseCase) Execute(ctx context.Context) StatusDTO {
	if err := u.query.PingMongo(ctx); err != nil {
		u.logger.ErrorContext(ctx, "health check failed", "error", err)
		return u.factory.Create(false)
	}
	return u.factory.Create(true)
}
