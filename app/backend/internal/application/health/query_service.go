// Package health is the use case that reports whether the backend can serve requests.
package health

import "context"

// QueryService reads the state of the stores the backend depends on.
// It is implemented in the infrastructure layer.
type QueryService interface {
	// PingMongo returns an error when MongoDB is unreachable.
	PingMongo(ctx context.Context) error
}
