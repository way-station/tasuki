package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// RunWorker retries the outbox entries that failed, once every first delay, until ctx is
// done (ADR-0016).
func (store *Outbox) RunWorker(ctx context.Context) {
	ticker := time.NewTicker(outboxFirstDelay)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			store.RetryDue(ctx)
		}
	}
}

// RetryDue runs the entries whose next attempt has come, a batch at a time. An entry that
// fails for the last time stays, and an error is logged.
func (store *Outbox) RetryDue(ctx context.Context) {
	for {
		rows, err := store.queries.ClaimDueOutboxEntries(ctx, sqlcgen.ClaimDueOutboxEntriesParams{
			Now:               timestamptz(store.now()),
			FirstDelaySeconds: int32(outboxFirstDelay / time.Second),
			MaxDelaySeconds:   int32(outboxMaxDelay / time.Second),
			MaxAttempts:       outboxMaxAttempts,
			BatchSize:         outboxBatchSize,
		})
		if err != nil {
			store.logger.ErrorContext(ctx, "claim outbox entries", "error", err)
			return
		}
		for _, row := range rows {
			err := store.run(ctx, row.ID, row.Steps)
			if err == nil {
				continue
			}
			if setErr := store.queries.SetOutboxEntryError(ctx, sqlcgen.SetOutboxEntryErrorParams{ID: row.ID, LastError: pgtype.Text{String: err.Error(), Valid: true}}); setErr != nil {
				store.logger.ErrorContext(ctx, "record the outbox entry error", "entry", row.ID, "error", setErr)
			}
			if row.Attempts >= outboxMaxAttempts {
				store.logger.ErrorContext(ctx, "outbox entry failed for the last time; it stays for someone to look at", "entry", row.ID, "attempts", row.Attempts, "error", err)
				continue
			}
			store.logger.WarnContext(ctx, "outbox entry failed again", "entry", row.ID, "attempts", row.Attempts, "error", err)
		}
		if len(rows) < outboxBatchSize {
			return
		}
	}
}
