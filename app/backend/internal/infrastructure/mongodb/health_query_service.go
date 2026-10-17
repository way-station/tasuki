package mongodb

import (
	"context"

	"github.com/way-station/tasuki/app/backend/internal/application/health"
)

// HealthQueryService implements health.QueryService with the MongoDB client.
type HealthQueryService struct {
	client *Client
}

var _ health.QueryService = (*HealthQueryService)(nil)

// NewHealthQueryService returns a HealthQueryService that uses client.
func NewHealthQueryService(client *Client) *HealthQueryService {
	return &HealthQueryService{client: client}
}

// PingMongo returns an error when MongoDB is unreachable.
func (q *HealthQueryService) PingMongo(ctx context.Context) error {
	return q.client.Ping(ctx)
}
