package postgres

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/health"
)

// HealthQueryService implements health.PostgresQueryService with the PostgreSQL client.
type HealthQueryService struct {
	client *Client
}

var _ health.PostgresQueryService = (*HealthQueryService)(nil)

// NewHealthQueryService returns a HealthQueryService that uses client.
func NewHealthQueryService(client *Client) *HealthQueryService {
	return &HealthQueryService{client: client}
}

// PingPostgres returns an error when PostgreSQL is unreachable.
func (q *HealthQueryService) PingPostgres(ctx context.Context) error {
	return q.client.Ping(ctx)
}
