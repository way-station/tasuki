package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/way-station/tasuki/app/backend/internal/domain/participation"
	"github.com/way-station/tasuki/app/backend/internal/domain/shared"
	"github.com/way-station/tasuki/app/backend/internal/infrastructure/postgres/sqlcgen"
)

// EntryRepository implements participation.EntryRepository with PostgreSQL. It runs in the
// transaction of a unit of work when there is one.
type EntryRepository struct {
	queries    *sqlcgen.Queries
	unitOfWork *UnitOfWork
	factory    participation.EntryFactory
}

var _ participation.EntryRepository = (*EntryRepository)(nil)

// NewEntryRepository returns an EntryRepository that uses client.
func NewEntryRepository(client *Client, factory participation.EntryFactory) *EntryRepository {
	return &EntryRepository{queries: sqlcgen.New(client.Pool()), unitOfWork: NewUnitOfWork(client), factory: factory}
}

// FindByID returns the entry with id and its members.
func (repository *EntryRepository) FindByID(ctx context.Context, id shared.EntryID) (*participation.Entry, error) {
	entryID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	row, err := queriesFor(ctx, repository.queries).GetEntry(ctx, entryID)
	if err != nil {
		return nil, notFound(err, "entry")
	}
	entries, err := repository.withMembers(ctx, []sqlcgen.ListEntriesByRaceRow{{ID: row.ID, RaceID: row.RaceID, Name: row.Name}})
	if err != nil {
		return nil, err
	}
	return entries[0], nil
}

// ListByRace returns the entries of the race and their members, oldest first.
func (repository *EntryRepository) ListByRace(ctx context.Context, raceID shared.RaceID) ([]*participation.Entry, error) {
	race, err := parseID(raceID)
	if err != nil {
		return nil, err
	}
	rows, err := queriesFor(ctx, repository.queries).ListEntriesByRace(ctx, race)
	if err != nil {
		return nil, err
	}
	return repository.withMembers(ctx, rows)
}

// withMembers reads the members of the entries and reconstructs them.
func (repository *EntryRepository) withMembers(ctx context.Context, rows []sqlcgen.ListEntriesByRaceRow) ([]*participation.Entry, error) {
	entryIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		entryIDs = append(entryIDs, row.ID)
	}
	memberRows, err := queriesFor(ctx, repository.queries).ListEntryMembers(ctx, entryIDs)
	if err != nil {
		return nil, err
	}
	members := map[uuid.UUID][]shared.MemberID{}
	for _, memberRow := range memberRows {
		members[memberRow.EntryID] = append(members[memberRow.EntryID], shared.MemberID(memberRow.RosterMemberID.String()))
	}
	entries := make([]*participation.Entry, 0, len(rows))
	for _, row := range rows {
		entries = append(entries, repository.factory.Reconstruct(shared.EntryID(row.ID.String()), shared.RaceID(row.RaceID.String()), row.Name, members[row.ID]))
	}
	return entries, nil
}

// Save stores a new entry, and makes the stored members the entry's members, in one
// transaction. Members who are removed have their legs of the running orders emptied
// (docs/design/database.md).
func (repository *EntryRepository) Save(ctx context.Context, entry *participation.Entry) error {
	entryID, err := parseID(entry.ID())
	if err != nil {
		return err
	}
	raceID, err := parseID(entry.RaceID())
	if err != nil {
		return err
	}
	memberIDs, err := parseIDs(entry.Members())
	if err != nil {
		return err
	}
	return repository.unitOfWork.Run(ctx, func(ctx context.Context) error {
		queries := queriesFor(ctx, repository.queries)
		if err := queries.SaveEntry(ctx, sqlcgen.SaveEntryParams{ID: entryID, RaceID: raceID, Name: entry.Name()}); err != nil {
			return err
		}
		if err := queries.RemoveEntryMembersExcept(ctx, sqlcgen.RemoveEntryMembersExceptParams{EntryID: entryID, Kept: memberIDs}); err != nil {
			return err
		}
		for _, memberID := range memberIDs {
			if err := queries.AddEntryMember(ctx, sqlcgen.AddEntryMemberParams{EntryID: entryID, RaceID: raceID, RosterMemberID: memberID}); err != nil {
				return err
			}
		}
		return nil
	})
}
