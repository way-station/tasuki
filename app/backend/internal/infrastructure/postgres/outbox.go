package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// The retries of an outbox entry (docs/design/database.md): 10 attempts in all, the
// retries 1, 2, 4, … minutes apart, at most 1 hour.
const (
	outboxMaxAttempts = 10
	outboxFirstDelay  = time.Minute
	outboxMaxDelay    = time.Hour
	outboxBatchSize   = 10
)

// OutboxStep tells the outbox how to run one kind of step. Each feature registers its
// steps when the backend is wired.
type OutboxStep struct {
	Kind string
	// Decode reads the step from the payload that the outbox stored as JSON.
	Decode func(payload []byte) (outbox.Step, error)
	// Run runs the step. It must have the same effect when it runs twice (ADR-0016).
	Run func(ctx context.Context, step outbox.Step) error
	// FinalPostgreSQL means the step writes to PostgreSQL. It must be the last step of an
	// entry, and Run runs in the transaction that deletes the entry, so it is applied once.
	FinalPostgreSQL bool
}

// Outbox implements outbox.Outbox with the outbox_entries table (ADR-0016). An entry is
// stored in the unit of work's transaction, its steps run right after the commit, and
// RunWorker retries the entries that failed.
type Outbox struct {
	queries    *sqlcgen.Queries
	unitOfWork *UnitOfWork
	ids        shared.IDGenerator
	steps      map[string]OutboxStep
	now        func() time.Time
	logger     *slog.Logger
}

var _ outbox.Outbox = (*Outbox)(nil)

// NewOutbox returns an Outbox that runs the registered steps.
func NewOutbox(client *Client, ids shared.IDGenerator, now func() time.Time, logger *slog.Logger, steps ...OutboxStep) (*Outbox, error) {
	registered := make(map[string]OutboxStep, len(steps))
	for _, step := range steps {
		if step.Kind == "" || step.Decode == nil || step.Run == nil {
			return nil, fmt.Errorf("outbox step %q needs a kind, Decode, and Run", step.Kind)
		}
		if _, ok := registered[step.Kind]; ok {
			return nil, fmt.Errorf("outbox step %q is registered twice", step.Kind)
		}
		registered[step.Kind] = step
	}
	return &Outbox{queries: sqlcgen.New(client.Pool()), unitOfWork: NewUnitOfWork(client), ids: ids, steps: registered, now: now, logger: logger}, nil
}

type storedStep struct {
	Kind    string          `json:"kind"`
	Payload json.RawMessage `json:"payload"`
}

// Add stores the entry in the transaction of ctx and runs its steps after the commit. It
// must run in a unit of work, and every step must be registered.
func (store *Outbox) Add(ctx context.Context, entry outbox.Entry) error {
	steps, err := store.encode(entry.Steps())
	if err != nil {
		return err
	}
	id, err := parseID(store.ids.NewID())
	if err != nil {
		return err
	}
	// The request runs the steps first; the retry worker waits for the first delay.
	params := sqlcgen.InsertOutboxEntryParams{ID: id, Steps: steps, NextAttemptAt: timestamptz(store.now().Add(outboxFirstDelay))}
	if !afterCommit(ctx, func(ctx context.Context) { store.runAfterCommit(ctx, id, steps) }) {
		return errors.New("outbox entries are added in a unit of work")
	}
	return queriesFor(ctx, store.queries).InsertOutboxEntry(ctx, params)
}

func (store *Outbox) encode(steps []outbox.Step) ([]byte, error) {
	stored := make([]storedStep, 0, len(steps))
	for i, step := range steps {
		registered, ok := store.steps[step.Kind()]
		if !ok {
			return nil, fmt.Errorf("outbox step %q is not registered", step.Kind())
		}
		if registered.FinalPostgreSQL && i != len(steps)-1 {
			return nil, fmt.Errorf("outbox step %q writes to PostgreSQL and must be the last step", step.Kind())
		}
		payload, err := json.Marshal(step)
		if err != nil {
			return nil, fmt.Errorf("encode outbox step %q: %w", step.Kind(), err)
		}
		stored = append(stored, storedStep{Kind: step.Kind(), Payload: payload})
	}
	return json.Marshal(stored)
}

// runAfterCommit runs the steps in the request. A failure is counted, and the retry worker
// tries again; the request has succeeded either way (ADR-0016).
func (store *Outbox) runAfterCommit(ctx context.Context, id uuid.UUID, steps []byte) {
	err := store.run(ctx, id, steps)
	if err == nil {
		return
	}
	store.logger.WarnContext(ctx, "outbox entry failed; the retry worker tries again", "entry", id, "error", err)
	failure := sqlcgen.FailOutboxEntryParams{ID: id, NextAttemptAt: timestamptz(store.now().Add(outboxFirstDelay)), LastError: pgtype.Text{String: err.Error(), Valid: true}}
	if err := store.queries.FailOutboxEntry(ctx, failure); err != nil {
		store.logger.ErrorContext(ctx, "count the failed outbox entry", "entry", id, "error", err)
	}
}

// run runs the steps in order, then deletes the entry in one transaction with the final
// PostgreSQL step, if there is one. When another attempt already deleted the entry, the
// final step is rolled back, so it is applied once.
func (store *Outbox) run(ctx context.Context, id uuid.UUID, encoded []byte) error {
	var stored []storedStep
	if err := json.Unmarshal(encoded, &stored); err != nil {
		return fmt.Errorf("decode outbox entry %s: %w", id, err)
	}
	var final *storedStep
	if last := len(stored) - 1; last >= 0 && store.steps[stored[last].Kind].FinalPostgreSQL {
		final, stored = &stored[last], stored[:last]
	}
	for _, step := range stored {
		if err := store.runStep(ctx, step); err != nil {
			return err
		}
	}
	err := store.unitOfWork.Run(ctx, func(ctx context.Context) error {
		if final != nil {
			if err := store.runStep(ctx, *final); err != nil {
				return err
			}
		}
		deleted, err := queriesFor(ctx, store.queries).DeleteOutboxEntry(ctx, id)
		if err != nil {
			return err
		}
		if deleted == 0 {
			return errOutboxEntryDone
		}
		return nil
	})
	if errors.Is(err, errOutboxEntryDone) {
		return nil
	}
	return err
}

var errOutboxEntryDone = errors.New("the outbox entry was already done")

func (store *Outbox) runStep(ctx context.Context, stored storedStep) error {
	registered, ok := store.steps[stored.Kind]
	if !ok {
		return fmt.Errorf("outbox step %q is not registered", stored.Kind)
	}
	step, err := registered.Decode(stored.Payload)
	if err != nil {
		return fmt.Errorf("decode outbox step %q: %w", stored.Kind, err)
	}
	if err := registered.Run(ctx, step); err != nil {
		return fmt.Errorf("outbox step %q: %w", stored.Kind, err)
	}
	return nil
}
