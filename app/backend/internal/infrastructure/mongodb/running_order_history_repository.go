package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/way-station/tasuki/app/backend/internal/domain/runningorder"
)

// RunningOrderHistoryRepository implements runningorder.HistoryRepository with the
// running_order_history collection (docs/design/database.md). MongoDB does not join the
// PostgreSQL unit of work, so an entry stays when the unit of work later rolls back.
type RunningOrderHistoryRepository struct {
	collection *mongo.Collection
}

var _ runningorder.HistoryRepository = (*RunningOrderHistoryRepository)(nil)

// NewRunningOrderHistoryRepository returns a RunningOrderHistoryRepository that uses client.
func NewRunningOrderHistoryRepository(client *Client) *RunningOrderHistoryRepository {
	return &RunningOrderHistoryRepository{collection: client.Database().Collection("running_order_history")}
}

type historyDocument struct {
	RunningOrderID  bson.Binary  `bson:"running_order_id"`
	RaceID          bson.Binary  `bson:"race_id"`
	EntryID         bson.Binary  `bson:"entry_id"`
	ChangedByUserID bson.Binary  `bson:"changed_by_user_id"`
	ChangedAt       time.Time    `bson:"changed_at"`
	Status          string       `bson:"status"`
	Legs            []historyLeg `bson:"legs"`
}

type historyLeg struct {
	LegNumber      int          `bson:"leg_number"`
	RosterMemberID *bson.Binary `bson:"roster_member_id"`
}

// Append stores the history entry. IDs are stored as BSON UUIDs.
func (repository *RunningOrderHistoryRepository) Append(ctx context.Context, historyEntry *runningorder.HistoryEntry) error {
	document := historyDocument{ChangedAt: historyEntry.ChangedAt(), Status: string(historyEntry.Status())}
	var err error
	if document.RunningOrderID, err = bsonUUID(string(historyEntry.RunningOrderID())); err != nil {
		return err
	}
	if document.RaceID, err = bsonUUID(string(historyEntry.RaceID())); err != nil {
		return err
	}
	if document.EntryID, err = bsonUUID(string(historyEntry.EntryID())); err != nil {
		return err
	}
	if document.ChangedByUserID, err = bsonUUID(string(historyEntry.ChangedBy())); err != nil {
		return err
	}
	document.Legs = make([]historyLeg, 0, len(historyEntry.Legs()))
	for _, leg := range historyEntry.Legs() {
		stored := historyLeg{LegNumber: leg.Number()}
		if runner, ok := leg.Runner(); ok {
			runnerID, err := bsonUUID(string(runner))
			if err != nil {
				return err
			}
			stored.RosterMemberID = &runnerID
		}
		document.Legs = append(document.Legs, stored)
	}
	if _, err := repository.collection.InsertOne(ctx, document); err != nil {
		return fmt.Errorf("append running order history: %w", err)
	}
	return nil
}

// bsonUUID converts an ID to a BSON UUID (binary subtype 4).
func bsonUUID(id string) (bson.Binary, error) {
	parsed, err := uuid.Parse(id)
	if err != nil {
		return bson.Binary{}, fmt.Errorf("invalid id %q: %w", id, err)
	}
	return bson.Binary{Subtype: bson.TypeBinaryUUID, Data: parsed[:]}, nil
}
