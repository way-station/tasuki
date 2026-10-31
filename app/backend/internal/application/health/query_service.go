// Package health is the use case that reports whether the backend can serve requests.
package health

import "context"

// MongoQueryService reads the state of MongoDB.
// It is implemented in the infrastructure layer.
type MongoQueryService interface {
	// PingMongo returns an error when MongoDB is unreachable.
	PingMongo(ctx context.Context) error
}

// PostgresQueryService reads the state of PostgreSQL.
// It is implemented in the infrastructure layer.
type PostgresQueryService interface {
	// PingPostgres returns an error when PostgreSQL is unreachable.
	PingPostgres(ctx context.Context) error
}
