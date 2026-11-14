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

type afterCommitKey struct{}

// afterCommitHooks are what runs after the transaction of a unit of work commits.
type afterCommitHooks struct {
	hooks []func(ctx context.Context)
}

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
// transaction, which commits or rolls back as a whole. After the commit, it runs what
// afterCommit registered, in the order it was registered.
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
	hooks := &afterCommitHooks{}
	transactionCtx := context.WithValue(context.WithValue(ctx, transactionKey{}, dbTransaction), afterCommitKey{}, hooks)
	if err := work(transactionCtx); err != nil {
		return err
	}
	if err := dbTransaction.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	for _, hook := range hooks.hooks {
		hook(ctx)
	}
	return nil
}

// afterCommit registers hook to run after the transaction of ctx commits, and reports false
// outside a unit of work. hook does not run when the transaction rolls back.
func afterCommit(ctx context.Context, hook func(ctx context.Context)) bool {
	hooks, ok := ctx.Value(afterCommitKey{}).(*afterCommitHooks)
	if !ok {
		return false
	}
	hooks.hooks = append(hooks.hooks, hook)
	return true
}

// queriesFor returns queries that run in the transaction of ctx, or queries itself
// outside a unit of work.
func queriesFor(ctx context.Context, queries *sqlcgen.Queries) *sqlcgen.Queries {
	if dbTransaction, ok := ctx.Value(transactionKey{}).(pgx.Tx); ok {
		return queries.WithTx(dbTransaction)
	}
	return queries
}
