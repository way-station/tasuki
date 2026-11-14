package postgres

import (
	"context"
	"fmt"

	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// RunningOrderRepository implements runningorder.Repository with PostgreSQL. It runs in the
// transaction of a unit of work when there is one.
type RunningOrderRepository struct {
	queries    *sqlcgen.Queries
	unitOfWork *UnitOfWork
	factory    runningorder.Factory
}

var _ runningorder.Repository = (*RunningOrderRepository)(nil)

// NewRunningOrderRepository returns a RunningOrderRepository that uses client.
func NewRunningOrderRepository(client *Client, factory runningorder.Factory) *RunningOrderRepository {
	return &RunningOrderRepository{queries: sqlcgen.New(client.Pool()), unitOfWork: NewUnitOfWork(client), factory: factory}
}

// FindByEntry returns the running order of the entry with its legs.
func (repository *RunningOrderRepository) FindByEntry(ctx context.Context, entryID shared.EntryID) (*runningorder.RunningOrder, error) {
	entry, err := parseID(entryID)
	if err != nil {
		return nil, err
	}
	queries := queriesFor(ctx, repository.queries)
	row, err := queries.GetRunningOrderByEntry(ctx, entry)
	if err != nil {
		return nil, notFound(err, "running order")
	}
	legRows, err := queries.ListRunningOrderLegs(ctx, row.ID)
	if err != nil {
		return nil, err
	}
	legs := make([]runningorder.Assignment, 0, len(legRows))
	for _, legRow := range legRows {
		leg := runningorder.Assignment{Number: int(legRow.LegNumber)}
		if legRow.RosterMemberID != nil {
			runner := shared.MemberID(legRow.RosterMemberID.String())
			leg.Runner = &runner
		}
		legs = append(legs, leg)
	}
	return repository.factory.Reconstruct(shared.RunningOrderID(row.ID.String()), entryID, runningorder.Status(row.Status), int(row.Version), legs), nil
}

// Save inserts a new running order, or updates it if the stored version is still the
// version it was read with, and replaces its legs, in one transaction. It returns
// shared.ErrVersionConflict otherwise.
func (repository *RunningOrderRepository) Save(ctx context.Context, saved *runningorder.RunningOrder) error {
	runningOrderID, err := parseID(saved.ID())
	if err != nil {
		return err
	}
	entryID, err := parseID(saved.EntryID())
	if err != nil {
		return err
	}
	version, err := toInt32(saved.Version(), "running order version")
	if err != nil {
		return err
	}
	loadedVersion, err := toInt32(saved.LoadedVersion(), "running order version")
	if err != nil {
		return err
	}
	legs := make([]sqlcgen.InsertRunningOrderLegParams, 0, len(saved.Legs()))
	for _, leg := range saved.Legs() {
		number, err := toInt32(leg.Number(), "leg number")
		if err != nil {
			return err
		}
		var runner *shared.MemberID
		if memberID, ok := leg.Runner(); ok {
			runner = &memberID
		}
		runnerID, err := parseOptionalID(runner)
		if err != nil {
			return err
		}
		legs = append(legs, sqlcgen.InsertRunningOrderLegParams{RunningOrderID: runningOrderID, EntryID: entryID, LegNumber: number, RosterMemberID: runnerID})
	}
	return repository.unitOfWork.Run(ctx, func(ctx context.Context) error {
		queries := queriesFor(ctx, repository.queries)
		affected, err := queries.SaveRunningOrder(ctx, sqlcgen.SaveRunningOrderParams{
			ID: runningOrderID, EntryID: entryID, Status: string(saved.Status()), Version: version, LoadedVersion: loadedVersion,
		})
		if err != nil {
			return err
		}
		if affected == 0 {
			return fmt.Errorf("running order %s: %w", saved.ID(), shared.ErrVersionConflict)
		}
		if err := queries.DeleteRunningOrderLegs(ctx, runningOrderID); err != nil {
			return err
		}
		for _, leg := range legs {
			if err := queries.InsertRunningOrderLeg(ctx, leg); err != nil {
				return err
			}
		}
		return nil
	})
}
