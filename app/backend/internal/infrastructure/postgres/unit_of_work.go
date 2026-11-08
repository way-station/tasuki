package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/way-station/tasuki/app/backend/internal/application/transaction"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

type transactionKey struct{}

// UnitOfWork implements transaction.UnitOfWork with one pgx transaction. The transaction
// is carried in the context, and repositories use it through queriesFor (ADR-0015).
type UnitOfWork struct {
	pool *pgxpool.Pool
}

var _ transaction.UnitOfWork = (*UnitOfWork)(nil)

// NewUnitOfWork returns a UnitOfWork that uses client.
func NewUnitOfWork(client *Client) *UnitOfWork {
	return &UnitOfWork{pool: client.Pool()}
}

// Run runs work in a new transaction. Inside another Run, work joins the outer
// transaction, which commits or rolls back as a whole.
func (unitOfWork *UnitOfWork) Run(ctx context.Context, work func(ctx context.Context) error) error {
	if _, ok := ctx.Value(transactionKey{}).(pgx.Tx); ok {
		return work(ctx)
	}
	dbTransaction, err := unitOfWork.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	// Rolling back after a commit does nothing.
	defer func() { _ = dbTransaction.Rollback(ctx) }()
	if err := work(context.WithValue(ctx, transactionKey{}, dbTransaction)); err != nil {
		return err
	}
	if err := dbTransaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// queriesFor returns queries that run in the transaction of ctx, or queries itself
// outside a unit of work.
func queriesFor(ctx context.Context, queries *sqlcgen.Queries) *sqlcgen.Queries {
	if dbTransaction, ok := ctx.Value(transactionKey{}).(pgx.Tx); ok {
		return queries.WithTx(dbTransaction)
	}
	return queries
}
