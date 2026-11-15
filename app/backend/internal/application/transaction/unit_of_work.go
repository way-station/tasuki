// Package transaction holds the unit of work that use cases run their writes in
// (ADR-0015, ADR-0016).
package transaction

import "context"

// UnitOfWork runs work in one PostgreSQL transaction. Repositories called with the
// context that work receives use that transaction. It is implemented in the
// infrastructure layer.
type UnitOfWork interface {
	// Run commits when work returns nil, and rolls back and returns the error otherwise.
	Run(ctx context.Context, work func(ctx context.Context) error) error
}
