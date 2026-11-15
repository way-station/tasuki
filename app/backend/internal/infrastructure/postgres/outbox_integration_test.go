//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/way-station/tasuki/app/backend/internal/application/outbox"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/random"
)

// countStep counts its runs in a memory step, as a MongoDB step would write a document.
type countStep struct {
	Name string `json:"name"`
}

func (countStep) Kind() string { return "test.count" }

// renameStep appends to the team name in the final PostgreSQL step.
type renameStep struct {
	Suffix string `json:"suffix"`
}

func (renameStep) Kind() string { return "test.rename" }

type outboxFixture struct {
	client   *Client
	store    *Outbox
	now      time.Time
	runs     map[string]int
	failures int // how many more times the count step fails
}

func newOutboxFixture(t *testing.T) *outboxFixture {
	t.Helper()
	fixture := &outboxFixture{client: newTestClient(t), now: time.Date(2026, 11, 10, 9, 0, 0, 0, time.UTC), runs: map[string]int{}}
	mustExec(t, fixture.client, "INSERT INTO teams (id, name) VALUES ($1, 'Team')", "01900000-0000-7000-8000-000000000001")
	mustExec(t, fixture.client, "DELETE FROM outbox_entries")
	count := OutboxStep{
		Kind: countStep{}.Kind(),
		Decode: func(payload []byte) (outbox.Step, error) {
			var step countStep
			return step, json.Unmarshal(payload, &step)
		},
		Run: func(_ context.Context, step outbox.Step) error {
			if fixture.failures > 0 {
				fixture.failures--
				return errors.New("mongo is down")
			}
			fixture.runs[step.(countStep).Name]++
			return nil
		},
	}
	rename := OutboxStep{
		Kind: renameStep{}.Kind(),
		Decode: func(payload []byte) (outbox.Step, error) {
			var step renameStep
			return step, json.Unmarshal(payload, &step)
		},
		Run: func(ctx context.Context, step outbox.Step) error {
			dbTransaction, ok := ctx.Value(transactionKey{}).(pgx.Tx)
			if !ok {
				return errors.New("the final step runs outside the transaction")
			}
			_, err := dbTransaction.Exec(ctx, "UPDATE teams SET name = name || $1", step.(renameStep).Suffix)
			return err
		},
		FinalPostgreSQL: true,
	}
	store, err := NewOutbox(fixture.client, random.IDGenerator{}, func() time.Time { return fixture.now }, slog.New(slog.NewTextHandler(io.Discard, nil)), count, rename)
	if err != nil {
		t.Fatal(err)
	}
	fixture.store = store
	return fixture
}

func (fixture *outboxFixture) add(t *testing.T, steps ...outbox.Step) error {
	t.Helper()
	entry, err := outbox.EntryFactory{}.New(steps...)
	if err != nil {
		t.Fatal(err)
	}
	return NewUnitOfWork(fixture.client).Run(context.Background(), func(ctx context.Context) error {
		return fixture.store.Add(ctx, entry)
	})
}

func (fixture *outboxFixture) teamName(t *testing.T) string {
	t.Helper()
	var name string
	if err := fixture.client.Pool().QueryRow(context.Background(), "SELECT name FROM teams").Scan(&name); err != nil {
		t.Fatal(err)
	}
	return name
}

type outboxRow struct {
	attempts      int
	nextAttemptAt time.Time
	lastError     *string
}

func (fixture *outboxFixture) rows(t *testing.T) []outboxRow {
	t.Helper()
	rows, err := fixture.client.Pool().Query(context.Background(), "SELECT attempts, next_attempt_at, last_error FROM outbox_entries")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var stored []outboxRow
	for rows.Next() {
		var row outboxRow
		if err := rows.Scan(&row.attempts, &row.nextAttemptAt, &row.lastError); err != nil {
			t.Fatal(err)
		}
		stored = append(stored, row)
	}
	return stored
}

func TestOutbox(t *testing.T) {
	t.Run("runs the steps after the commit and deletes the entry with the final step", func(t *testing.T) {
		fixture := newOutboxFixture(t)
		if err := fixture.add(t, countStep{Name: "history"}, renameStep{Suffix: "!"}); err != nil {
			t.Fatalf("add: %v", err)
		}
		if fixture.runs["history"] != 1 || fixture.teamName(t) != "Team!" || len(fixture.rows(t)) != 0 {
			t.Errorf("runs %v, team %q, rows %v, want one run, Team!, and no row", fixture.runs, fixture.teamName(t), fixture.rows(t))
		}
	})

	t.Run("keeps nothing when the unit of work rolls back", func(t *testing.T) {
		fixture := newOutboxFixture(t)
		entry, err := outbox.EntryFactory{}.New(countStep{Name: "history"})
		if err != nil {
			t.Fatal(err)
		}
		failure := errors.New("roll back")
		err = NewUnitOfWork(fixture.client).Run(context.Background(), func(ctx context.Context) error {
			if err := fixture.store.Add(ctx, entry); err != nil {
				return err
			}
			return failure
		})
		if !errors.Is(err, failure) || fixture.runs["history"] != 0 || len(fixture.rows(t)) != 0 {
			t.Errorf("Run error = %v, runs %v, rows %v, want nothing", err, fixture.runs, fixture.rows(t))
		}
	})

	t.Run("leaves a failed entry for the retry worker, which finishes it", func(t *testing.T) {
		fixture := newOutboxFixture(t)
		fixture.failures = 1
		if err := fixture.add(t, countStep{Name: "history"}, renameStep{Suffix: "!"}); err != nil {
			t.Fatalf("add: %v", err)
		}
		rows := fixture.rows(t)
		if len(rows) != 1 || rows[0].attempts != 1 || !rows[0].nextAttemptAt.Equal(fixture.now.Add(time.Minute)) || rows[0].lastError == nil || fixture.teamName(t) != "Team" {
			t.Fatalf("rows %+v, team %q, want one failed attempt due in a minute and the team unchanged", rows, fixture.teamName(t))
		}

		fixture.store.RetryDue(context.Background())
		if fixture.runs["history"] != 0 || len(fixture.rows(t)) != 1 {
			t.Fatalf("runs %v, rows %v, want the entry not due yet", fixture.runs, fixture.rows(t))
		}
		fixture.now = fixture.now.Add(time.Minute)
		fixture.store.RetryDue(context.Background())
		if fixture.runs["history"] != 1 || fixture.teamName(t) != "Team!" || len(fixture.rows(t)) != 0 {
			t.Errorf("runs %v, team %q, rows %v, want the entry finished once", fixture.runs, fixture.teamName(t), fixture.rows(t))
		}
	})

	t.Run("applies the final step once when an entry runs twice", func(t *testing.T) {
		fixture := newOutboxFixture(t)
		steps, err := fixture.store.encode([]outbox.Step{countStep{Name: "history"}, renameStep{Suffix: "!"}})
		if err != nil {
			t.Fatal(err)
		}
		id := random.IDGenerator{}.NewID()
		mustExec(t, fixture.client, "INSERT INTO outbox_entries (id, steps, next_attempt_at) VALUES ($1, $2, $3)", id, steps, fixture.now)
		entryID, err := parseID(id)
		if err != nil {
			t.Fatal(err)
		}
		for range 2 {
			if err := fixture.store.run(context.Background(), entryID, steps); err != nil {
				t.Fatalf("run: %v", err)
			}
		}
		if fixture.runs["history"] != 2 || fixture.teamName(t) != "Team!" {
			t.Errorf("runs %v, team %q, want the memory step twice and the final step once", fixture.runs, fixture.teamName(t))
		}
	})

	t.Run("stops after 10 attempts and keeps the entry", func(t *testing.T) {
		fixture := newOutboxFixture(t)
		fixture.failures = 100
		if err := fixture.add(t, countStep{Name: "history"}); err != nil {
			t.Fatalf("add: %v", err)
		}
		var delays []time.Duration
		for range 12 {
			before := fixture.rows(t)[0]
			fixture.now = before.nextAttemptAt
			fixture.store.RetryDue(context.Background())
			if after := fixture.rows(t)[0]; after.attempts > before.attempts {
				delays = append(delays, after.nextAttemptAt.Sub(fixture.now))
			}
		}
		rows := fixture.rows(t)
		if len(rows) != 1 || rows[0].attempts != 10 {
			t.Fatalf("rows %+v, want the entry kept after 10 attempts", rows)
		}
		want := []time.Duration{2 * time.Minute, 4 * time.Minute, 8 * time.Minute, 16 * time.Minute, 32 * time.Minute, time.Hour, time.Hour, time.Hour, time.Hour}
		if len(delays) != len(want) {
			t.Fatalf("delays %v, want %v", delays, want)
		}
		for i := range want {
			if delays[i] != want[i] {
				t.Errorf("delay %d = %v, want %v", i+2, delays[i], want[i])
			}
		}
	})

	t.Run("refuses entries it cannot run", func(t *testing.T) {
		fixture := newOutboxFixture(t)
		if err := fixture.add(t, renameStep{Suffix: "!"}, countStep{Name: "history"}); err == nil {
			t.Error("add with the PostgreSQL step first succeeded")
		}
		if err := fixture.add(t, unknownStep{}); err == nil {
			t.Error("add with an unregistered step succeeded")
		}
		entry, err := outbox.EntryFactory{}.New(countStep{Name: "history"})
		if err != nil {
			t.Fatal(err)
		}
		if err := fixture.store.Add(context.Background(), entry); err == nil {
			t.Error("Add outside a unit of work succeeded")
		}
		if len(fixture.rows(t)) != 0 {
			t.Errorf("rows %v, want none", fixture.rows(t))
		}
	})
}

type unknownStep struct{}

func (unknownStep) Kind() string { return "test.unknown" }
